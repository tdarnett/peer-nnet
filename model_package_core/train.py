# USAGE
# python train.py

# import packages

import json
import os
import time
from importlib.resources import path
from pathlib import Path
from types import ModuleType
from typing import Dict, List

import matplotlib.pyplot as plt
import torch
from sklearn.model_selection import train_test_split
from torch.utils.data import DataLoader, Subset
from tqdm import tqdm

from .pytorch_model.dataset import ProcessedDataset
from .pytorch_model.early_stopping import EarlyStopping
from .pytorch_model.model import Net


def load_model(
    base_model: Net,
    metadata_and_weights: Dict[str, Dict],
    total_samples: int,
    config: ModuleType
) -> Net:
    """Load the latest version of the local host model, and, if applicable, update weights using peer networks.

    :param base_model: Model object to update weights on
    :param metadata_and_weights: Dictionary containing the metadata and weight files for peer networks
    :param total_samples: Total number of training samples used across all incoming peer networks
    :param config: configuration file
    :return: Model with new weights, ready for training
    """
    # load the weights for the local model
    print('[INFO] Load existing model weights & biases...')
    base_model.load_state_dict(torch.load(config.MODEL_PATH))

    # if there are updated weights from peers available
    if metadata_and_weights:
        print('[INFO] Found new model weights & biases. Updating existing parameters...')
        # Update the local model weights
        for p_out in base_model.parameters():
            p_out.data = torch.nn.Parameter((config.NUMBER_OF_TRAIN_SAMPLES / total_samples) * p_out)

        # instantiate a temporary model object
        temp_model = Net(
            input_size=config.INPUT_SIZE,
            hidden_units=config.HIDDEN_UNITS,
            number_of_classes=config.NUMBER_OF_CLASSES
        ).to(config.DEVICE)

        # for all of the new peer weights, update the local weights via weighted sums of training samples
        for model_key in metadata_and_weights:
            temp_model.load_state_dict(torch.load(metadata_and_weights[model_key]['weights']))
            for p_out, p_in in zip(base_model.parameters(), temp_model.parameters()):
                p_out.data = torch.nn.Parameter(p_out + (metadata_and_weights[model_key]['number_of_samples'] / total_samples) * p_in)

    return base_model


def plot_and_save_training_loss(history: Dict[str, List[float]], plot_path: str):
    """Plot and save the training loss.

    :param history: dictionary with train and validation loss history
    :param plot_path: path to save plots
    """
    # plot the training loss
    plt.style.use('ggplot')
    plt.figure()
    plt.plot(history['train_loss'], label='train_loss')
    plt.plot(history['val_loss'], label='val_loss')
    plt.title('Training Loss on Dataset')
    plt.xlabel('Epoch #')
    plt.ylabel('Loss')
    plt.legend(loc='lower left')
    plt.savefig(plot_path)


def save_metadata(metadata_path: str, number_of_training_samples: int):
    """Save the metadata from training run in JSON file format.

    :param metadata_path: Path to metadata JSON file
    :param number_of_training_samples: Number of training samples used to train model
    """
    # update model version in metadata
    if Path(metadata_path).exists():
        with open(metadata_path, 'r') as metadata_file:
            metadata_dict = json.load(metadata_file)
            model_version = metadata_dict['version'] + 1
    else:
        # first time a model is being training
        model_version = 1

    # save metadata
    metadata_dict = {
        'version' : model_version,
        'sample_size' : number_of_training_samples,
        'last_updated' : time.ctime()
    }
    metadata_string = json.dumps(metadata_dict)
    with open(metadata_path, 'w') as outfile:
        outfile.write(metadata_string)


def train(metadata_and_weights: Dict[str, Dict], config: ModuleType):
    # create the dataset
    dataset = ProcessedDataset(
        images_path=config.TRAIN_IMAGE_DATA_PATH,
        labels_path=config.TRAIN_LABEL_DATA_PATH,
        number_of_samples=config.NUMBER_OF_TRAIN_SAMPLES
    )

    # create data indices for training and validation splits
    train_indices, val_indices, _, _ = train_test_split(
        range(config.NUMBER_OF_TRAIN_SAMPLES),
        dataset.labels,
        stratify=dataset.labels,
        test_size=config.VAL_SIZE,
        random_state=config.RANDOM_SEED
    )

    # generate subset based on indices
    train_split = Subset(dataset, train_indices)
    val_split = Subset(dataset, val_indices)

    print(f'[INFO] found {len(train_indices)} examples in the training set...')
    print(f'[INFO] found {len(val_indices)} examples in the validation set...')

    # create the training and validation data loaders
    train_loader = DataLoader(
        dataset=train_split,
        shuffle=True,
        batch_size=config.BATCH_SIZE,
        pin_memory=config.PIN_MEMORY,
        num_workers=os.cpu_count()
    )
    val_loader = DataLoader(
        dataset=val_split,
        batch_size=config.BATCH_SIZE,
        pin_memory=config.PIN_MEMORY,
        num_workers=os.cpu_count()
    )

    # initialize a model
    model = Net(
        input_size=config.INPUT_SIZE,
        hidden_units=config.HIDDEN_UNITS,
        number_of_classes=config.NUMBER_OF_CLASSES,
        lr=config.INIT_LR
    ).to(config.DEVICE)

    # load and update existing weights & biases
    if Path(config.MODEL_PATH).exists():
        total_samples = sum([v['number_of_samples'] for v in metadata_and_weights.values()])
        total_samples += config.NUMBER_OF_TRAIN_SAMPLES
        model = load_model(
            base_model=model,
            metadata_and_weights=metadata_and_weights,
            total_samples=total_samples,
            config=config
        )

    # calculate steps per epoch for training and validation set
    train_steps = len(train_indices) // config.BATCH_SIZE
    val_steps = len(val_indices) // config.BATCH_SIZE

    # initialize a dictionary to store training history
    history = {'train_loss': [], 'val_loss': []}

    # instantiate early stopping
    early_stopping = EarlyStopping(tolerance=config.TOLERANCE, min_delta=config.MIN_DELTA)

    # loop over epochs
    print('[INFO] training the network...')
    start_time = time.time()

    for epoch in tqdm(range(config.NUM_EPOCHS)):
        # set the model in training mode
        model.train()

        # initialize the total training and validation loss
        total_train_loss = 0
        total_val_loss = 0

        # loop over the training set
        for (x_batch, y_batch) in train_loader:
            # send the input to the device
            (x_batch, y_batch) = (x_batch.to(config.DEVICE), y_batch.to(config.DEVICE))

            # perform a forward pass and calculate the training loss
            pred = model(x_batch.float())
            loss = model.loss_fn(pred, y_batch)

            # first, zero out any previously accumulated gradients, then
            # perform backpropagation, and then update model parameters
            model.optimizer.zero_grad()
            loss.backward()
            model.optimizer.step()

            # add the loss to the total training loss so far
            total_train_loss += loss

        # switch off autograd
        with torch.no_grad():
            # set model in evaluation mode
            model.eval()

            # loop over the validation set
            for (x_batch, y_batch) in val_loader:
                # send the input to the device
                (x_batch, y_batch) = (x_batch.to(config.DEVICE), y_batch.to(config.DEVICE))

                # make the predictions and calculate the validation loss
                pred = model(x_batch.float())

                # add the loss to the total validation loss so far
                total_val_loss += model.loss_fn(pred, y_batch.long())

        # calculate the average training and validation accuracy
        average_train_loss = total_train_loss / train_steps
        average_val_loss = total_val_loss / val_steps

        # update the training history
        history['train_loss'].append(average_train_loss.cpu().detach().numpy())
        history['val_loss'].append(average_val_loss.cpu().detach().numpy())

        # print the model training and validation information
        print('[INFO] EPOCH: {}/{}'.format(epoch + 1, config.NUM_EPOCHS))
        print('Train loss: {:.3f}, Validation loss: {:.3f}'.format(average_train_loss, average_val_loss))

        # early stopping
        early_stopping(average_train_loss, average_val_loss)
        if early_stopping.early_stop:
            print('[INFO] Stopping at Epoch:', epoch + 1)
            break

    # display the total time needed to perform the training
    end_time = time.time()
    print('[INFO] total time taken to train the model: {:.2f}s'.format(end_time - start_time))

    # serialize the model to disk
    torch.save(model.state_dict(), config.MODEL_PATH)

    # plot the training loss
    plot_and_save_training_loss(history=history, plot_path=config.PLOT_PATH)

    # save metadata for training run
    save_metadata(metadata_path=config.METADATA_PATH, number_of_training_samples=config.NUMBER_OF_TRAIN_SAMPLES)
