from psx.runtime import BaseView, Element, el, escape, fragment
from typing import List, Optional, Dict
class ComplexView(BaseView):
    def __init__(self, title: str, items: List[str]=[], metadata: Optional[Dict[str, str]]=None, *args, **kwargs):
        super().__init__()
        self.title = title
        self.items = items
        self.metadata = metadata
        self.args = args
        self.kwargs = kwargs

    def _render(self) -> Element:
        return el("div", [el("h1", escape(self.title)), el("p", f"Items count:{escape(len(self.items))}"), "", el("p", f"Args:{escape(len(self.args))}"), el("p", f"Kwargs:{escape(len(self.kwargs))}")])

