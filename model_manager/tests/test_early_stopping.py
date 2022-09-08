def test_early_stopping_class_init(create_early_stopping):
    # GIVEN a set tolerance and minimum delta
    early_stopping = create_early_stopping

    # THEN counter is initialized to 0
    assert early_stopping.counter == 0

    # AND early_stop is false
    assert early_stopping.early_stop == False


def test_early_stopping_class_exceed_tolerance(create_early_stopping):
    # GIVEN a set tolerance and minimum delta
    early_stopping = create_early_stopping

    # WHEN number of loss deltas exceeds min delta tolerance
    for _ in range(3):
        early_stopping(1, 3)

    # THEN early_stop is true
    assert early_stopping.early_stop == True
