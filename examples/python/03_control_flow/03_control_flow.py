def control_flow(n):
    if n < 0:
        return "negative"
    elif n == 0:
        return "zero"
    else:
        result = []
        while n > 0:
            result.append(n)
            n -= 1
        return result


for i in range(3, -2, -1):
    print(control_flow(i)) 