from topple.psx import BaseView, Element, el, escape, fragment, raw
class FStringExpressions(BaseView):
    def __init__(self, name: str, items: list, total: float):
        super().__init__()
        self.name = name
        self.items = items
        self.total = total

    def _render(self) -> Element:
        _root_children_1000 = []
        discount = 0.1
        _div_children_2000 = []
        _div_children_2000.append(el("h1", f"Welcome, {escape(self.name)}!"))
        _div_children_2000.append(el("p", f"You have {escape(len(self.items))} items in your cart"))
        _div_children_2000.append(el("p", f"Total: ${escape(self.total)}"))
        _div_children_2000.append(el("p", f"Discount: {escape(discount)}"))
        _div_children_2000.append(el("p", f"Final total: ${escape(self.total * (1 - discount))}"))
        for (i, item) in enumerate(self.items):
            _div_children_2000.append(el("div", f"{escape(i + 1)}. {escape(item)}"))
        _div_children_2000.append(el("div", escape(f"Processing {"completed" if len(self.items) > 0 else "pending"}")))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

