from topple.psx import BaseView, Element, el, escape, fragment
from topple.psx import raw
class CodeExample(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        psx_code = "<div class=\\\"card\\\">Hello, World!</div>"
        html_code = "<strong>Bold text</strong>"
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "Testing raw() Function"))
        _div_children_2000.append(el("h2", "Without raw() - should be escaped"))
        _div_children_2000.append(el("pre", el("code", escape(psx_code))))
        _div_children_2000.append(el("h2", "With raw() - should render literally"))
        _div_children_2000.append(el("pre", el("code", escape(raw(psx_code)))))
        _div_children_2000.append(el("h2", "Security test - normal content still escaped"))
        _div_children_2000.append(el("p", escape("<script>alert('XSS')</script>")))
        _div_children_2000.append(el("h2", "Raw HTML renders"))
        _div_children_2000.append(el("div", escape(raw(html_code))))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

