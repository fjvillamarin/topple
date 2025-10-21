from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class ConditionalView(BaseView):
    def __init__(self, user_type: str, is_admin: bool=False):
        super().__init__()
        self.user_type = user_type
        self.is_admin = is_admin

    def _render(self) -> Element:
        return el("div", ["", ""])

