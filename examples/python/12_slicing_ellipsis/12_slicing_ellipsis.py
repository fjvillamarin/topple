try:
    import numpy as np
except ImportError:
    np = None  # placeholder if numpy not installed


if np is not None:
    a = np.arange(27).reshape(3, 3, 3)
    print(a[..., 1])
else:
    # Fallback to show ellipsis token in slice without numpy
    sample = ((0, 1), (2, 3), (4, 5))
    print(sample[...]) 