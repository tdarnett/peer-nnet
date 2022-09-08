from pathlib import Path

PEER_MODEL_PATH = Path('.', 'shared', 'peers', 'models')

# must be in sync with networking app filename constants!
WEIGHT_FILENAME = 'weights.h5'
METADATA_FILENAME = 'metadata.json'

TRAIN_IMAGE_DATA_PATH = Path('model_manager/tests/fixtures/train_images.npy')
TRAIN_LABEL_DATA_PATH = Path('model_manager/tests/fixtures/train_labels.npy')
