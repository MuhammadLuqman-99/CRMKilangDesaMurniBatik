interface HeaderProps {
    isMenuOpen: boolean;
    setIsMenuOpen: (value: boolean) => void;
}

export default function Header({ isMenuOpen, setIsMenuOpen }: HeaderProps) {
    return (
        <header className="header">
            <div className="container header-content">
                <a href="/" className="logo">
                    <span className="logo-icon">K</span>
                    <span className="logo-text">Kilang Batik CRM</span>
                </a>

                <nav className={`nav ${isMenuOpen ? 'nav-open' : ''}`}>
                    <a href="#features" className="nav-link">Features</a>
                    <a href="#pricing" className="nav-link">Pricing</a>
                    <a href="#testimonials" className="nav-link">Testimonials</a>
                    <a href="#faq" className="nav-link">FAQ</a>
                    <a href="#contact" className="nav-link">Contact</a>
                </nav>

                <div className="header-actions">
                    <a href="/login" className="btn btn-ghost">Log In</a>
                    <a href="/register" className="btn btn-primary">Start Free Trial</a>
                </div>

                <button
                    className="mobile-menu-btn"
                    onClick={() => setIsMenuOpen(!isMenuOpen)}
                    aria-label="Toggle menu"
                >
                    <span className={`hamburger ${isMenuOpen ? 'open' : ''}`}></span>
                </button>
            </div>
        </header>
    );
}
