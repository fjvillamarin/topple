from topple.psx import BaseView, Element, el, escape, fragment
class ConditionalView(BaseView):
    def __init__(self, user_type: str, is_admin: bool=False):
        super().__init__()
        self.user_type = user_type
        self.is_admin = is_admin

    def _render(self) -> Element:
        return el("div", ["", ""])

