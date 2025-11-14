from topple.psx import BaseView, Element, el, escape, fragment
from functools import wraps
def cache_result(func):
    cache = {}
    @wraps(func)
    def wrapper(*args, **kwargs):
        key = str(args) + str(kwargs)
        if key not in cache:
            cache[key] = func(*args, **kwargs)
        return cache[key]

    return wrapper

@cache_result
def expensive_operation(n: int) -> int:
    return n ** 2 + n * 10

def logged(cls):
    original_init = cls.__init__
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
        processor = DataProcessor("MyProcessor")
        results = processor.process(self.values)
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "Decorator Demo"))
        _div_children_2000.append(el("p", f"Processor: {escape(processor.name)}"))
        _div_children_2000.append(el("p", f"Input values: {escape(self.values)}"))
        _div_children_2000.append(el("p", f"Results: {escape(results)}"))
        for (i, result) in enumerate(results):
            _div_children_2000.append(el("div", f"f"Result {escape(i)}: {escape(result)}""))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

