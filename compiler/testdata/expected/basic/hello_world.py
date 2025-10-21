from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class HelloWorld(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", "Hello World")

