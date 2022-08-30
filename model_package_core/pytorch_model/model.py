# very simple model to train on the MNIST dataset
# this file is used to hold the architecture of any neural net

import torch


class Net(torch.nn.Module):
    def __init__(self, input_size, hidden_units, number_of_classes, lr=0.001):
        super(Net, self).__init__()
        # initialize the layer list
        all_layers = []
        # create the layers
        for hidden_unit in hidden_units:
            layer = torch.nn.Linear(input_size, hidden_unit)
            all_layers.append(layer)
            all_layers.append(torch.nn.ReLU())
            input_size = hidden_unit
        all_layers.append(torch.nn.Linear(hidden_units[-1], number_of_classes))
        self.module_list = torch.nn.ModuleList(all_layers)
        self.loss_fn = torch.nn.CrossEntropyLoss()
        self.optimizer = torch.optim.Adam(self.parameters(), lr=lr)

    def forward(self, x):
        # apply each layer to input
        for f in self.module_list:
            x = f(x)
        return x
