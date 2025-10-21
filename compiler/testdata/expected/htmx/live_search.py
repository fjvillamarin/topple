from topple.psx import BaseView
from topple.psx import Element
from topple.psx import el
from topple.psx import escape
from topple.psx import fragment
class SearchInterface(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [el("h1", "Live Search"), el("input", "", {"type": "search", "name": "q", "placeholder": "Search products...", "hx-get": "/api/search", "hx-trigger": "keyup changed delay:500ms, search", "hx-target": "#search-results", "hx-indicator": ".search-spinner"}), el("div", el("span", "Searching..."), {"class": "search-spinner", "style": "display:none;"}), el("div", el("p", "Start typing to search...", {"class": "placeholder"}), {"id": "search-results"}), el("div", el("select", [el("option", "All Categories", {"value": ""}), el("option", "Electronics", {"value": "electronics"}), el("option", "Books", {"value": "books"}), el("option", "Clothing", {"value": "clothing"})], {"name": "category", "hx-get": "/api/search", "hx-trigger": "change", "hx-target": "#search-results", "hx-include": "[name='q']"}), {"class": "filters"})], {"class": "search-container"})

