from topple.psx import BaseView, Element, el, escape, fragment
from functools import wraps
def cache_result(func):
    _chain_tmp_1 = {}
    cache = _chain_tmp_1
    @wraps(func)
    def wrapper(*args, **kwargs):
        _chain_tmp_2 = str(args) + str(kwargs)
        key = _chain_tmp_2
        if key not in cache:
            _chain_tmp_3 = func(*args, **kwargs)
            cache[key] = _chain_tmp_3
        return cache[key]

    return wrapper

@cache_result
def expensive_operation(n: int) -> int:
    return n ** 2 + n * 10

def logged(cls):
    _chain_tmp_4 = cls.__init__
    original_init = _chain_tmp_4
    def new_init(self, *args, **kwargs):
        print(f"Creating {cls.__name__}")
        original_init(self, *args, **kwargs)

    cls.__init__ = new_init
    return cls

@logged
class DataProcessor:
    def __init__(self, name: str):
        self.name = name

    def process(self, data: list) -> list:
        return [expensive_operation(x) for x in data if x > 0]

class DecoratorDemo(BaseView):
    def __init__(self, values: list):
        super().__init__()
        self.values = values

    def _render(self) -> Element:
        _root_children_1000 = []
        _chain_tmp_5 = DataProcessor("MyProcessor")
        processor = _chain_tmp_5
        _chain_tmp_6 = processor.process(self.values)
        results = _chain_tmp_6
        _root_children_1000.append(el("div", [el("h1", "Decorator Demo"), el("p", f"Processor:{escape(processor.name)}"), el("p", f"Input values:{escape(self.values)}"), el("p", f"Results:{escape(results)}"), ""]))
        return fragment(_root_children_1000)

