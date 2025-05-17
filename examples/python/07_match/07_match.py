def classify(val):
    match val:
        case 0:
            return "zero"
        case 1 | 2 | 3:
            return "small"
        case [x, y]:
            return f"pair {x} {y}"
        case {"key": value}:
            return f"dict with key={value}"
        case _:
            return "something else"


if __name__ == "__main__":
    samples = [0, 2, [10, 20], {"key": 99}, None]
    for s in samples:
        print(classify(s)) 