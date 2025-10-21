from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class MultiRoot(BaseView):
    def __init__(self, title: str, content: str):
        super().__init__()
        self.title = title
        self.content = content

    def _render(self) -> Element:
        el("h1", escape(self.title))
        el("p", escape(self.content))
        el("div", el("span", "Additional content"))

