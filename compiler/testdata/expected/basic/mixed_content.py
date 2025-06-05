from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
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
        _view_children_1000 = []
        current_time = datetime.now()
        _view_children_1000.append(el("div", [el("h1", "Counter App"), el("p", f"Current count:{escape(increment())}"), el("p", f"Time:{escape(current_time)}")]))
        return fragment(_view_children_1000)

