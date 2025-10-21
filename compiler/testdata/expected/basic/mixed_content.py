from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
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
        _chain_tmp_1 = datetime.now()
        current_time = _chain_tmp_1
        el("div", [el("h1", "Counter App"), el("p", f"Current count:{escape(increment())}"), el("p", f"Time:{escape(current_time)}")])

