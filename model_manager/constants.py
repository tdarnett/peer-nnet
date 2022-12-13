"""
Constants specific to *running* the model manager.

Any model configuration settings should live in the pytorch config.py!
"""
from pathlib import Path

PEER_MODEL_PATH = Path('.', 'shared', 'peers', 'models')

# must be in sync with networking app filename constants!
WEIGHT_FILENAME = 'weights.h5'
METADATA_FILENAME = 'metadata.json'
