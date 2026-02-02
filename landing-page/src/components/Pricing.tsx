import { useState } from 'react';

const plans = [
    {
        name: 'Starter',
        description: 'Perfect for small teams getting started',
        monthlyPrice: 49,
        yearlyPrice: 39,
        features: [
            'Up to 5 users',
            '1,000 contacts',
            'Basic pipeline management',
            'Email support',
            'Mobile access',
            'Basic reporting',
        ],
        highlighted: false,
        cta: 'Start Free Trial',
    },
    {
        name: 'Professional',
        description: 'For growing businesses with advanced needs',
        monthlyPrice: 99,
        yearlyPrice: 79,
        features: [
            'Up to 25 users',
            '10,000 contacts',
            'Advanced pipelines',
            'Priority support',
            'API access',
            'Custom reports',
            'Workflow automation',
            'Lead scoring',
        ],
        highlighted: true,
        cta: 'Start Free Trial',
        badge: 'Most Popular',
    },
    {
        name: 'Enterprise',
        description: 'For large organizations with custom requirements',
        monthlyPrice: 199,
        yearlyPrice: 159,
        features: [
            'Unlimited users',
            'Unlimited contacts',
            'Multiple pipelines',
            'Dedicated support',
            'Custom integrations',
            'Advanced analytics',
            'SSO & 2FA',
            'SLA guarantee',
            'Custom training',
        ],
        highlighted: false,
        cta: 'Contact Sales',
    },
];

export default function Pricing() {
    const [isYearly, setIsYearly] = useState(true);

    return (
        <section id="pricing" className="pricing">
            <div className="container">
                <div className="section-header">
                    <span className="section-badge">Pricing</span>
                    <h2 className="section-title">
                        Simple, <span className="gradient-text">transparent pricing</span>
                    </h2>
                    <p className="section-description">
                        Choose the plan that's right for your business. All plans include a
                        14-day free trial.
                    </p>
                </div>

                <div className="pricing-toggle">
                    <span className={!isYearly ? 'active' : ''}>Monthly</span>
                    <button
                        className={`toggle-switch ${isYearly ? 'yearly' : ''}`}
                        onClick={() => setIsYearly(!isYearly)}
                        aria-label="Toggle billing period"
                    >
                        <span className="toggle-knob"></span>
                    </button>
                    <span className={isYearly ? 'active' : ''}>
                        Yearly <span className="save-badge">Save 20%</span>
                    </span>
                </div>

                <div className="pricing-grid">
                    {plans.map((plan, index) => (
                        <div
                            key={index}
                            className={`pricing-card ${plan.highlighted ? 'highlighted' : ''}`}
                        >
                            {plan.badge && <span className="plan-badge">{plan.badge}</span>}
                            <h3 className="plan-name">{plan.name}</h3>
                            <p className="plan-description">{plan.description}</p>
                            <div className="plan-price">
                                <span className="currency">RM</span>
                                <span className="amount">
                                    {isYearly ? plan.yearlyPrice : plan.monthlyPrice}
                                </span>
                                <span className="period">/user/month</span>
                            </div>
                            {isYearly && (
                                <p className="billing-note">Billed annually</p>
                            )}
                            <ul className="plan-features">
                                {plan.features.map((feature, i) => (
                                    <li key={i}>
                                        <svg className="check-icon" viewBox="0 0 24 24" fill="currentColor">
                                            <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
                                        </svg>
                                        {feature}
                                    </li>
                                ))}
                            </ul>
                            <a
                                href="/register"
                                className={`btn ${plan.highlighted ? 'btn-primary' : 'btn-outline'} btn-block`}
                            >
                                {plan.cta}
                            </a>
                        </div>
                    ))}
                </div>

                <p className="pricing-note">
                    All prices in Malaysian Ringgit (MYR). Need a custom plan?{' '}
                    <a href="#contact">Contact us</a>
                </p>
            </div>
        </section>
    );
}
