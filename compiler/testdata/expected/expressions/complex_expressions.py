from psx.runtime import BaseView, Element, el, escape, fragment
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
        _chain_tmp_1 = sum((item.get("price", 0) for item in items))
        total_value = _chain_tmp_1
        _root_children_1000.append(el("div", [el("h1", escape(self.user.get("name", "Anonymous").title())), el("p", f"Items:{escape(len(self.items))}({escape(get_status(len(self.items)))})"), el("p", f"Total:{escape(format_currency(total_value))}"), el("p", f"Average:{escape(format_currency(total_value / len(self.items)) if self.items else "N/A")}"), el("div", "", {"class": escape(f"status-{get_status(len(items))}")}), el("ul", ""), ""]))
        return fragment(_root_children_1000)

