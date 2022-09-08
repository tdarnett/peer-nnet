# NOTE: this should be a generic test for any model architecture
# we need to ensure that a loss function (loss_fn) and optimizer (optimizer)
# is instantiated for model training and that a forward pass is functional
def test_nn_model_class_init(create_model):
    # GIVEN a model specification
    model = create_model

    # THEN loss function is present
    assert model.loss_fn

    # AND optimizer is present
    assert model.optimizer


def test_nn_model_class_forward_pass(create_model):
    # GIVEN a model specification
    model = create_model

    # WHEN forward pass is called
    import torch
    temp_input = torch.rand(784)
    pred = model.forward(temp_input)

    # THEN output of size 10 is generated
    assert len(pred) == 10
