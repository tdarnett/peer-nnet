import json
import os
import time
from pathlib import Path

from sqlitedict import SqliteDict

from model_package_core.pytorch_model import config
from model_package_core.train import train


def create_default_db():
    # initialize DB
    db = SqliteDict('peer_metadata.sqlite')

    # add database elements
    db['peer_1'] = {'version': 1, 'sample_size': 640, 'last_updated': 'Thu Aug 25 21:39:49 2022'}
    db['peer_2'] = {'version': 1, 'sample_size': 320, 'last_updated': 'Fri Aug 26 20:30:00 2022'}
    db['peer_3'] = {'version': 2, 'sample_size': 1280, 'last_updated': 'Thu Aug 25 22:37:16 2022'}

    # commit to save the objects
    db.commit()

    # close database connection
    db.close()


def create_peer_files(id: int, version: int, sample_size: int):

    # set file paths
    PEER_MODEL_PATH = Path(f'peers/models/peer_{id}/')
    WEIGHT_FILENAME = Path('weights.h5')
    METADATA_FILENAME = Path('metadata.json')

    # create peer directory
    os.makedirs(PEER_MODEL_PATH, exist_ok=True)

    if not (PEER_MODEL_PATH / WEIGHT_FILENAME).exists():
        # add metadata.json
        metadata = {
            'version': version,
            'sample_size': sample_size,
            'last_updated': time.ctime()
        }
        metadata_string = json.dumps(metadata)
        with open(PEER_MODEL_PATH / METADATA_FILENAME, 'w') as outfile:
            outfile.write(metadata_string)

        # train model
        config.BASE_OUTPUT = PEER_MODEL_PATH
        config.MODEL_PATH = PEER_MODEL_PATH / WEIGHT_FILENAME
        config.METADATA_PATH = PEER_MODEL_PATH / METADATA_FILENAME
        config.PLOT_PATH = PEER_MODEL_PATH / Path('plot.png')
        config.NUMBER_OF_TRAIN_SAMPLES = sample_size
        train(metadata_and_weights={}, config=config)


if __name__ == '__main__':
    create_default_db()
    # train base / local model if non already
    if not Path(config.MODEL_PATH).exists():
        train(metadata_and_weights={}, config=config)
    create_peer_files(id=1, version=1, sample_size=640)
    create_peer_files(id=2, version=1, sample_size=320)
    create_peer_files(id=3, version=1, sample_size=1280)
