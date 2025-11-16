"""
FastAPI server for Topple/PSX landing page
"""

from fastapi import FastAPI
from fastapi.responses import HTMLResponse

# Import the compiled PSX views
from examples.homepage.pages import LandingPage

app = FastAPI(
    title="Topple Landing Page",
    description="Landing page for Topple/PSX - JSX-like syntax for Python",
    version="0.1.0"
)


@app.get("/", response_class=HTMLResponse)
async def home():
    """
    Render the landing page using PSX view components
    """
    page = LandingPage()
    return page.render()


@app.get("/health")
async def health():
    """
    Health check endpoint
    """
    return {"status": "healthy", "message": "Topple landing page is running"}
