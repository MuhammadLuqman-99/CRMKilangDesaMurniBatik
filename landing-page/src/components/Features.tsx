const features = [
    {
        icon: '📊',
        title: 'Visual Sales Pipeline',
        description:
            'Drag-and-drop Kanban boards to manage deals through every stage. Get a clear view of your entire sales process.',
    },
    {
        icon: '👥',
        title: 'Lead Management',
        description:
            'Capture, qualify, and convert leads with AI-powered scoring. Never miss a hot prospect again.',
    },
    {
        icon: '🏢',
        title: 'Customer 360° View',
        description:
            'See the complete picture of every customer relationship with contacts, activities, and deal history.',
    },
    {
        icon: '📈',
        title: 'Analytics & Reporting',
        description:
            'Make data-driven decisions with real-time dashboards, forecasting, and custom reports.',
    },
    {
        icon: '🔔',
        title: 'Smart Notifications',
        description:
            'Stay on top of important updates with intelligent alerts for deals, tasks, and customer activities.',
    },
    {
        icon: '🔐',
        title: 'Enterprise Security',
        description:
            'Bank-grade security with SSO, 2FA, role-based access control, and full audit logging.',
    },
    {
        icon: '📱',
        title: 'Mobile Ready',
        description:
            'Access your CRM anywhere with our responsive design. Works perfectly on any device.',
    },
    {
        icon: '🔌',
        title: 'API & Integrations',
        description:
            'Connect with your favorite tools via our REST API and pre-built integrations.',
    },
];

export default function Features() {
    return (
        <section id="features" className="features">
            <div className="container">
                <div className="section-header">
                    <span className="section-badge">Features</span>
                    <h2 className="section-title">
                        Everything you need to <span className="gradient-text">close more deals</span>
                    </h2>
                    <p className="section-description">
                        Powerful features designed to streamline your sales process and help
                        your team work smarter, not harder.
                    </p>
                </div>

                <div className="features-grid">
                    {features.map((feature, index) => (
                        <div key={index} className="feature-card">
                            <div className="feature-icon">{feature.icon}</div>
                            <h3 className="feature-title">{feature.title}</h3>
                            <p className="feature-description">{feature.description}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}
