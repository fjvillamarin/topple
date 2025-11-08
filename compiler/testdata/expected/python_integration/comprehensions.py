from topple.psx import BaseView, Element, el, escape, fragment
class Comprehensions(BaseView):
    def __init__(self, numbers: list, items: list):
        super().__init__()
        self.numbers = numbers
        self.items = items

    def _render(self) -> Element:
        _root_children_1000 = []
        squares = [x ** 2 for x in self.numbers if x % 2 == 0]
        item_names = {item["id"]: item["name"] for item in self.items if "name" in item}
        unique_lengths = {len(item["name"]) for item in self.items if "name" in item}
        _div_children_2000 = []
        _div_children_2000.append(el("h2", "Comprehensions Demo"))
        _div_children_2000.append(el("h3", "Even squares:"))
        _ul_children_3000 = []
        for square in squares:
            _ul_children_3000.append(el("li", escape(square)))
        _div_children_2000.append(el("ul", _ul_children_3000))
        _div_children_2000.append(el("h3", "Item mapping:"))
        _ul_children_4000 = []
        for (item_id, name) in item_names.items():
            _ul_children_4000.append(el("li", f"{escape(item_id)}:{escape(name)}"))
        _div_children_2000.append(el("ul", _ul_children_4000))
        _div_children_2000.append(el("h3", "Unique name lengths:"))
        _div_children_2000.append(el("p", f"f"Lengths:{escape(sorted(unique_lengths))}""))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

