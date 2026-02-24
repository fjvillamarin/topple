from topple.psx import BaseView, Element, el, escape, fragment, raw
class BooleanAttributes(BaseView):
    def __init__(self, is_editable: bool=False, is_required: bool=True):
        super().__init__()
        self.is_editable = is_editable
        self.is_required = is_required

    def _render(self) -> Element:
        _root_children_1000 = []
        _form_children_2000 = []
        _form_children_2000.append(el("input", "", {"type": "text", "readonly": escape(not self.is_editable), "required": escape(self.is_required), "disabled": escape(False), "autofocus": escape(True)}))
        _form_children_2000.append(el("input", "", {"type": "checkbox", "checked": escape(self.is_editable)}))
        _form_children_2000.append(el("button", escape("Submit"), {"type": "submit", "disabled": escape(not self.is_editable)}))
        _root_children_1000.append(el("form", _form_children_2000))
        return fragment(_root_children_1000)

