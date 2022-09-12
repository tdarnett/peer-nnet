import json
from pathlib import Path

from sqlitedict import SqliteDict

from model_package_core.constants import WEIGHT_FILENAME, PEER_MODEL_PATH, METADATA_FILENAME


class ModelMetadataSync:
    """
    Scans a directory of peer model data, identifies new data,
    and triggers a local model training process on the new data.
    """
    peer_models: Path
    weights_filename: str
    metadata_filename: str

    def __init__(self, db: SqliteDict, peer_models: Path = PEER_MODEL_PATH,
                 weights_filename: str = WEIGHT_FILENAME, metadata_filename: str = METADATA_FILENAME):
        self.db = db
        self.metadata_filename = metadata_filename
        self.weights_filename = weights_filename
        self.peer_models = peer_models

    def run(self):
        models_to_train = self._models_to_train()
        if models_to_train:
            from model_package_core.train import train
            from pytorch_model import config

            train(metadata_and_weights=models_to_train, config=config)

    def _models_to_train(self) -> dict[str, dict]:
        """
        Generates a mapping of peer ID and it's model weights and metadata, for any new peer model
        version.
        """
        models_to_train = {}
        for peer_path in self.peer_models.iterdir():
            assert peer_path.is_dir()
            peer_id = peer_path.name
            # open metadata to inspect it's model version
            peer_metadata_file = Path(peer_path, self.metadata_filename)
            with open(peer_metadata_file, 'r') as metadata_file:
                latest_metadata = json.load(metadata_file)

            # compare with last seen local datastore version
            db_metadata = self.db.get(peer_id)
            if db_metadata is None or latest_metadata['version'] > db_metadata['version']:
                # update datastore
                self.db[peer_id] = latest_metadata
                self.db.commit()

                # add weight file pointers and metadata to output
                models_to_train[peer_id] = dict(
                    weights=Path(peer_path, self.weights_filename),
                    number_of_samples=latest_metadata['sample_size']
                )

        return models_to_train

