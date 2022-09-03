import json

from model_package_core.constants import METADATA_FILENAME, WEIGHT_FILENAME
from model_package_core.sync import ModelMetadataSync


def test_new_db_syncs_all_peers(peers_db, peer_models_path):
    # GIVEN an empty peers db and a peers models path that has 3 peers
    assert len(peers_db) == 0

    # WHEN we generate the models to train
    sync_commander = ModelMetadataSync(db=peers_db, peer_models=peer_models_path)
    models_to_train = sync_commander._models_to_train()

    # THEN
    peers_in_peer_path = []
    for peer_path in peer_models_path.iterdir():
        peer_id = peer_path.name
        peers_in_peer_path.append(peer_id)

        # the DB should be synced with the latest values
        assert peers_db[peer_id] is not None
        with open(peer_path / METADATA_FILENAME, 'r') as f:
            assert peers_db[peer_id] == json.load(f)

        # AND the models to train should have the correct data
        assert peers_db[peer_id]['sample_size'] == models_to_train[peer_id]['number_of_samples']
        assert models_to_train[peer_id]['weights'] == peer_path / WEIGHT_FILENAME

    # AND the models to train dict and DB has all the new peers
    assert len(models_to_train) == len(peers_in_peer_path) == len(peers_db)


def test_syncs_new_peer(peers_db, peer_models_path):
    # TODO
    pass


def test_syncs_existing_peer_new_model_version(peers_db, peer_models_path):
    # TODO
    pass
