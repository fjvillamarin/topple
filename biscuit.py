from typing import Optional, Dict, Callable, Any, Protocol
from io import StringIO

class Writer:
    def __init__(self):
        self._buffer = StringIO()
    
    def write(self, content: str) -> None:
        self._buffer.write(content)
    
    def to_string(self) -> str:
        return self._buffer.getvalue()

class View(Protocol):
    def render(self, writer: Optional[Writer] = None) -> str:
        ...
    
    def render_to(self, writer: Writer) -> None:
        ...
    
    def __str__(self) -> str:
        ...

class AnonymousView(View):
    def __init__(self, render_fn: Callable[[Writer], Any]):
        self._render_fn = render_fn
    
    def render(self, writer: Writer = Writer()) -> str:
        self.render_to(writer)
        return writer.to_string()
    
    def render_to(self, writer: Writer) -> None:
        self._render_fn(writer)
    
    def __str__(self) -> str:
        return self.render()
    
class Slots:
    def __init__(self):
        self.slots: Dict[str, View] = {}
    
    def add(self, name: str, view: View) -> None:
        self.slots[name] = view
    
    def get(self, name: str, default: Optional[View] = None) -> View:
        if default is None:
            default = AnonymousView(lambda writer: None)
        return self.slots.get(name, default)
    
    def __contains__(self, name: str) -> bool:
        return name in self.slots
    
    def __getitem__(self, name: str) -> View:
        return self.slots[name]
    
    def __setitem__(self, name: str, value: View) -> None:
        self.slots[name] = value