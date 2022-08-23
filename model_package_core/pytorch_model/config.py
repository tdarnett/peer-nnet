# configuration file for training the NN model

# import packages
import os

import torch

# base path of the data
DATA_PATH = os.path.join('model_package_core', 'data')

# train data path
TRAIN_IMAGE_DATA_PATH = os.path.join(DATA_PATH, 'train/MNIST/raw/train-images-idx3-ubyte')
TRAIN_LABEL_DATA_PATH = os.path.join(DATA_PATH, 'train/MNIST/raw/train-labels-idx1-ubyte')

# test data path
TEST_IMAGE_DATA_PATH = os.path.join(DATA_PATH, 'test/MNIST/raw/t10k-images-idx3-ubyte')
TEST_LABEL_DATA_PATH = os.path.join(DATA_PATH, 'test/MNIST/raw/t10k-labels-idx1-ubyte')

# define the number of samples to use
NUMBER_OF_TRAIN_SAMPLES = 320

# define the train / validation split ratio
VAL_SIZE = 0.2
SHUFFLE_DATASET = True
RANDOM_SEED = 42

# determine the device to be used for training and evaluation
DEVICE = 'cuda' if torch.cuda.is_available() else 'cpu'

# determine if we will be pinning memory during data loading
PIN_MEMORY = True if DEVICE == 'cuda' else False

# define the number of hidden units
HIDDEN_UNITS = [32, 16]
INPUT_SIZE = 784
NUMBER_OF_CLASSES = 10

# initialize learning rate, number of epochs to train for, and the batch size
INIT_LR = 0.001
NUM_EPOCHS = 3
BATCH_SIZE = 64

# define the number of samples to plot during model evaluation
SAMPLES_TO_PLOT = 5

# define the path to the base output directory
BASE_OUTPUT = os.path.join('model_package_core', 'output')

# define the path to the output serialized model and model training plot
MODEL_PATH = os.path.join(BASE_OUTPUT, 'model_weights_updated_weights.h5')
PLOT_PATH = os.path.sep.join([BASE_OUTPUT, 'plot.png'])

# experimenting with model logic
# TODO can remove these
TOTAL_SAMPLES = 960
METADATA_WEIGHTS = {
    'model_2': {'weights': os.path.join(BASE_OUTPUT, 'model_weights_2.h5'), 'number_of_samples': 320},
    'model_3': {'weights': os.path.join(BASE_OUTPUT, 'model_weights_3.h5'), 'number_of_samples': 320}
}
