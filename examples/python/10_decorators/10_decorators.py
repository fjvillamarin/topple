def uppercase(fn):
    def inner(*args, **kwargs):
        result = fn(*args, **kwargs)
        return result.upper()
    return inner


@uppercase
def greet():
    return "hello"


if __name__ == "__main__":
    print(greet()) 