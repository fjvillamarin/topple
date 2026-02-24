from topple.psx import BaseView, Element, el, escape, fragment, raw
from datetime import datetime
count = 0
def increment():
    global count
    count = count + 1
    return count

class Counter(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        current_time = datetime.now()
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "Counter App"))
        _div_children_2000.append(el("p", f"Current count: {escape(increment())}"))
        _div_children_2000.append(el("p", f"Time: {escape(current_time)}"))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

