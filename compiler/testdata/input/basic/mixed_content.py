from topple.psx import BaseView, Element, el, escape, fragment
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
        current_time = datetime.now()
        return el("div", [el("h1", "Counter App"), el("p", f"Current count:{escape(increment())}"), el("p", f"Time:{escape(current_time)}")])

