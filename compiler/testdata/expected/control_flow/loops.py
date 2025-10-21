from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class LoopView(BaseView):
    def __init__(self, items: list, max_count: int=10):
        super().__init__()
        self.items = items
        self.max_count = max_count

    def _render(self) -> Element:
        return el("div", [el("h2", "Items List"), el("ul", ""), "", ""])

