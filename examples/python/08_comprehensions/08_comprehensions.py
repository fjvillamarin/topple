squares = [x * x for x in range(10)]
evans = {x for x in range(20) if x % 2 == 0}
pairs = {(x, y): x * y for x in range(3) for y in range(3)}

gen = (x ** 2 for x in range(5))

if __name__ == "__main__":
    print(squares)
    print(sorted(evans))
    print(pairs)
    print(list(gen)) 