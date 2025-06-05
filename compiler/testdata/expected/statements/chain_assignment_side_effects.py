counter = 0
def increment_and_return():
    global counter
    counter = counter + 1
    return counter

_chain_tmp_1 = increment_and_return()
c = _chain_tmp_1
b = _chain_tmp_1
a = _chain_tmp_1
_chain_tmp_2 = increment_and_return() + increment_and_return()
z = _chain_tmp_2
y = _chain_tmp_2
x = _chain_tmp_2
_chain_tmp_3 = print("This should only print once") or 42
s = _chain_tmp_3
r = _chain_tmp_3
q = _chain_tmp_3
p = _chain_tmp_3
