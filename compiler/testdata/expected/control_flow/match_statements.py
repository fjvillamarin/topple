from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class MatchView(BaseView):
    def __init__(self, status: str, data: dict):
        super().__init__()
        self.status = status
        self.data = data

    def _render(self) -> Element:
        return el("div", "")

