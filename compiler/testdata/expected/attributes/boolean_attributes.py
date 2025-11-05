from topple.psx import BaseView, Element, el, escape, fragment
class BooleanAttributes(BaseView):
    def __init__(self, is_editable: bool=False, is_required: bool=True):
        super().__init__()
        self.is_editable = is_editable
        self.is_required = is_required

    def _render(self) -> Element:
        return el("form", [el("input", "", {"type": "text", "readonly": escape(not self.is_editable), "required": escape(self.is_required), "disabled": escape(False), "autofocus": escape(True)}), el("input", "", {"type": "checkbox", "checked": escape(self.is_editable)}), el("button", "", {"type": "submit", "disabled": escape(not self.is_editable)})])

