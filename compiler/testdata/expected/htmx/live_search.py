from topple.psx import BaseView, Element, el, escape, fragment
class SearchInterface(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "Live Search"))
        _div_children_2000.append(el("input", "", {"type": "search", "name": "q", "placeholder": "Search products...", "hx-get": "/api/search", "hx-trigger": "keyup changed delay:500ms, search", "hx-target": "#search-results", "hx-indicator": ".search-spinner"}))
        _div_children_2000.append(el("div", el("span", "Searching..."), {"class": "search-spinner", "style": "display:none;"}))
        _div_children_2000.append(el("div", el("p", "Start typing to search...", {"class": "placeholder"}), {"id": "search-results"}))
        _div_children_3000 = []
        _select_children_4000 = []
        _select_children_4000.append(el("option", "All Categories", {"value": ""}))
        _select_children_4000.append(el("option", "Electronics", {"value": "electronics"}))
        _select_children_4000.append(el("option", "Books", {"value": "books"}))
        _select_children_4000.append(el("option", "Clothing", {"value": "clothing"}))
        _div_children_3000.append(el("select", _select_children_4000, {"name": "category", "hx-get": "/api/search", "hx-trigger": "change", "hx-target": "#search-results", "hx-include": "[name='q']"}))
        _div_children_2000.append(el("div", _div_children_3000, {"class": "filters"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "search-container"}))
        return fragment(_root_children_1000)

