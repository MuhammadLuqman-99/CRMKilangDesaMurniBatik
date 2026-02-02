export default function Hero() {
    return (
        <section className="hero">
            <div className="container hero-content">
                <div className="hero-text">
                    <span className="hero-badge">New: AI-Powered Lead Scoring</span>
                    <h1 className="hero-title">
                        Grow Your Business with
                        <span className="gradient-text"> Smart CRM</span>
                    </h1>
                    <p className="hero-description">
                        The all-in-one CRM platform designed for Malaysian businesses.
                        Manage leads, close deals, and build lasting customer relationships
                        with ease.
                    </p>
                    <div className="hero-cta">
                        <a href="/register" className="btn btn-primary btn-lg">
                            Start Free 14-Day Trial
                        </a>
                        <a href="#demo" className="btn btn-outline btn-lg">
                            <svg className="icon" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M8 5v14l11-7z" />
                            </svg>
                            Watch Demo
                        </a>
                    </div>
                    <div className="hero-stats">
                        <div className="stat">
                            <span className="stat-value">500+</span>
                            <span className="stat-label">Active Users</span>
                        </div>
                        <div className="stat">
                            <span className="stat-value">RM 50M+</span>
                            <span className="stat-label">Deals Closed</span>
                        </div>
                        <div className="stat">
                            <span className="stat-value">98%</span>
                            <span className="stat-label">Customer Satisfaction</span>
                        </div>
                    </div>
                </div>
                <div className="hero-image">
                    <div className="dashboard-preview">
                        <div className="preview-header">
                            <div className="preview-dots">
                                <span></span>
                                <span></span>
                                <span></span>
                            </div>
                        </div>
                        <div className="preview-content">
                            <div className="preview-sidebar"></div>
                            <div className="preview-main">
                                <div className="preview-cards">
                                    <div className="preview-card"></div>
                                    <div className="preview-card"></div>
                                    <div className="preview-card"></div>
                                </div>
                                <div className="preview-chart"></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div className="hero-gradient"></div>
        </section>
    );
}
