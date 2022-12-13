import json
from pathlib import Path

import numpy as np
from model_manager.constants import METADATA_FILENAME, WEIGHT_FILENAME
from model_manager.pytorch_model import config
from model_manager.train import TrainingLoop


def test_prepare_data_functionality(train_label_path, train_image_path):
    # GIVEN a local model to train with updated config
    config.TRAIN_LABEL_DATA_PATH = train_label_path
    config.TRAIN_IMAGE_DATA_PATH = train_image_path
    config.NUMBER_OF_TRAIN_SAMPLES = 80
    config.BATCH_SIZE = 16

    # WHEN data is prepared
    trainer = TrainingLoop(metadata_and_weights={}, config=config)
    trainer.prepare_data()

    # THEN training and validation data is loaded into DataLoader objects
    assert len(trainer.train_loader.dataset) == (1 - config.VAL_SIZE) * config.NUMBER_OF_TRAIN_SAMPLES
    assert len(trainer.val_loader.dataset) == config.VAL_SIZE * config.NUMBER_OF_TRAIN_SAMPLES


def test_initialize_model_functionality():
    # GIVEN a local model to train with updated config
    config.NUMBER_OF_TRAIN_SAMPLES = 80

    # WHEN a model is initialize with parameters from config
    trainer = TrainingLoop(metadata_and_weights={}, config=config)
    trainer.initialize_model()

    # THEN model weights have been initialized
    assert (trainer.model.state_dict())


def test_plot_and_save_model_history(local_model_path):
    # GIVEN a default model history
    plot_path = local_model_path / Path('plot.png')
    history = {'train_loss': np.arange(10), 'val_loss': np.arange(10)}

    # WHEN model training loss is plotted and saved to disk
    trainer = TrainingLoop(metadata_and_weights={}, config=config)
    trainer._plot_and_save_training_loss(history=history, plot_path=plot_path)

    # THEN training history is plotted
    assert plot_path.exists()


def test_save_metadata_functionality(local_model_path):
    # GIVEN metadata configuration
    metadata_path = local_model_path/ METADATA_FILENAME
    number_of_training_samples = 80

    # WHEN metadata is saved
    trainer = TrainingLoop(metadata_and_weights={}, config=config)
    trainer._save_metadata(metadata_path=metadata_path, number_of_training_samples=number_of_training_samples)

    # THEN model version has been updated (instantiated)
    with open(metadata_path, "r") as json_file:
        metadata = json.load(json_file)
    assert metadata['version'] == 1
    assert metadata['sample_size'] == number_of_training_samples


def test_train_process(local_model_path, train_label_path, train_image_path):
    # GIVEN a local model to train with updated config
    config.MODEL_PATH = local_model_path / WEIGHT_FILENAME
    config.METADATA_PATH = local_model_path/ METADATA_FILENAME
    config.PLOT_PATH = local_model_path / Path('plot.png')
    config.TRAIN_LABEL_DATA_PATH = train_label_path
    config.TRAIN_IMAGE_DATA_PATH = train_image_path
    config.NUMBER_OF_TRAIN_SAMPLES = 80
    config.NUM_EPOCHS = 1
    config.BATCH_SIZE = 16

    # WHEN a model is trained
    trainer = TrainingLoop(metadata_and_weights={}, config=config)
    trainer.initialize_model()
    trainer.prepare_data()
    trainer.run_training()

    # THEN model version has been updated (instantiated)
    with open(local_model_path / METADATA_FILENAME, "r") as json_file:
        metadata = json.load(json_file)
    assert metadata['version'] == 1
    assert metadata['sample_size'] == config.NUMBER_OF_TRAIN_SAMPLES

    # AND model weights are saved
    assert config.MODEL_PATH.exists()

    # AND training history is plotted
    assert config.PLOT_PATH.exists()
