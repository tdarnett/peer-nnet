# USAGE
# python predict.py

# import the necessary packages
import matplotlib.pyplot as plt
import numpy as np
import torch
from mlxtend.data import loadlocal_mnist
from sklearn.metrics import classification_report

from pytorch_model import config
from pytorch_model.model import Net


def prepare_plot(original_image: np.ndarray, original_label: np.uint8, predicted_label: np.ndarray):
    """Plot MNIST image with true and predicted label.

    :param original_image: Single MNIST image to plot
    :param original_label: Image true label
    :param predicted_label: Image predicted label
    """
    # reshape the array into 28 x 28 array (2-dimensional array)
    original_image = original_image.reshape((28, 28))

    # plot
    plt.title(f'Actual Label: {original_label}\nPredicted Label: {predicted_label}')
    plt.imshow(original_image, cmap='gray')
    plt.show()


def make_predictions(model: Net, image: np.ndarray, label: np.uint8):
    """Predict image label and plot results.

    :param model: Neural Network Model
    :param image: Single MNIST image to make prediction on
    :param label: Image true label
    """
    # set model to evaluation mode
    model.eval()

    # turn off gradient tracking
    with torch.no_grad():
        # make the prediction
        pred = model(torch.Tensor(image / 255.).to(config.DEVICE))
        pred = np.argmax(pred).cpu().detach().numpy()

    prepare_plot(image, label, pred)


def evaluate_performance(model: Net, images: np.ndarray, labels: np.ndarray):
    """Evaluate model performance on the test data.

    :param model: Neural Network Model
    :param images: Test MNIST images
    :param labels: Image labels
    """
    # set model to evaluation mode
    model.eval()

    # turn off gradient tracking
    with torch.no_grad():
        # make the prediction
        predictions = model(torch.Tensor(images / 255.).to(config.DEVICE))
        predictions = np.argmax(predictions, 1)

    print(f'[INFO] Evaluating model on {images.shape[0]} test samples...\n')
    print(classification_report(y_true=labels, y_pred=predictions, zero_division=1))


if __name__ == '__main__':

    # load the images and randomly select 10
    print('[INFO] loading the test image paths...')
    images, labels = loadlocal_mnist(images_path=config.TEST_IMAGE_DATA_PATH, labels_path=config.TEST_LABEL_DATA_PATH)
    idx = np.random.randint(0, images.shape[0], config.SAMPLES_TO_PLOT)
    selected_samples = zip(images[idx], labels[idx])

    # instantiate model and load weights from disk and flash it to the current device
    print('[INFO] load the model...')
    model = Net(
        input_size=config.INPUT_SIZE,
        hidden_units=config.HIDDEN_UNITS,
        number_of_classes=config.NUMBER_OF_CLASSES
    ).to(config.DEVICE)
    model.load_state_dict(torch.load(config.MODEL_PATH))

    # evaluate model performance on test data
    evaluate_performance(model, images, labels)

    # iterate over the randomly selected test images
    for image, label in selected_samples:
        # make predictions and visualize the results
        make_predictions(model, image, label)
