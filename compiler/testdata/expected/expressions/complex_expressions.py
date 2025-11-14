from topple.psx import BaseView, Element, el, escape, fragment
def format_currency(amount):
    return f"${amount:,.2f}"

def get_status(count):
    if count == 0:
        return "empty"
    else:
        if count < 5:
            return "low"
        else:
            return "good"

class ComplexExpressions(BaseView):
    def __init__(self, items: list, user: dict):
        super().__init__()
        self.items = items
        self.user = user

    def _render(self) -> Element:
        _root_children_1000 = []
        total_value = sum((item.get("price", 0) for item in self.items))
        _div_children_2000 = []
        _div_children_2000.append(el("h1", escape(self.user.get("name", "Anonymous").title())))
        _div_children_2000.append(el("p", f"Items: {escape(len(self.items))} ({escape(get_status(len(self.items)))})"))
        _div_children_2000.append(el("p", f"Total: {escape(format_currency(total_value))}"))
        _div_children_2000.append(el("p", f"Average: {escape(format_currency(total_value / len(self.items)) if self.items else "N/A")}"))
        _div_children_2000.append(el("div", "", {"class": escape(f"status-{get_status(len(self.items))}")}))
        _ul_children_3000 = []
        for item in self.items[:3]:
            _ul_children_3000.append(el("li", ""))
        _div_children_2000.append(el("ul", _ul_children_3000))
        if len(self.items) > 3:
            _div_children_2000.append(el("p", f"f"... and {escape(len(self.items) - 3)} more items""))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

