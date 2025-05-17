def greet(name: str = "World") -> str:
    return f"Hello, {name}!"


adder = lambda x, y=10: x + y


def variadic(*args, **kwargs):
    return args, kwargs 