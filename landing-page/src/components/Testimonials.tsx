const testimonials = [
    {
        quote:
            "Since implementing Kilang Batik CRM, our sales team's productivity has increased by 40%. The visual pipeline makes it so easy to track deals.",
        name: 'Ahmad Rahman',
        title: 'Sales Director',
        company: 'TechMalaysia Sdn Bhd',
        avatar: 'AR',
    },
    {
        quote:
            "The best CRM I've used for the Malaysian market. Local support, MYR pricing, and features that actually make sense for how we do business here.",
        name: 'Siti Aminah',
        title: 'CEO',
        company: 'Kreative Solutions',
        avatar: 'SA',
    },
    {
        quote:
            "We migrated from Salesforce and couldn't be happier. It's more affordable, easier to use, and the local support team is incredibly responsive.",
        name: 'Raj Kumar',
        title: 'Operations Manager',
        company: 'Global Trade MY',
        avatar: 'RK',
    },
    {
        quote:
            "The lead scoring feature has transformed how we prioritize prospects. We're closing 30% more deals with the same team size.",
        name: 'Lee Wei Ming',
        title: 'Sales Manager',
        company: 'PropertyPro Malaysia',
        avatar: 'LW',
    },
    {
        quote:
            "Finally, a CRM that understands Malaysian business needs. The multi-currency support and local integrations are game changers.",
        name: 'Nurul Huda',
        title: 'Business Development',
        company: 'Fintech Ventures',
        avatar: 'NH',
    },
    {
        quote:
            "Setup was incredibly easy. We were up and running in less than a day, and the team picked it up without any formal training.",
        name: 'Chen Mei Ling',
        title: 'Founder',
        company: 'Startup Hub KL',
        avatar: 'CM',
    },
];

export default function Testimonials() {
    return (
        <section id="testimonials" className="testimonials">
            <div className="container">
                <div className="section-header">
                    <span className="section-badge">Testimonials</span>
                    <h2 className="section-title">
                        Loved by <span className="gradient-text">businesses across Malaysia</span>
                    </h2>
                    <p className="section-description">
                        Don't just take our word for it. Here's what our customers have to say.
                    </p>
                </div>

                <div className="testimonials-grid">
                    {testimonials.map((testimonial, index) => (
                        <div key={index} className="testimonial-card">
                            <div className="quote-icon">"</div>
                            <p className="testimonial-quote">{testimonial.quote}</p>
                            <div className="testimonial-author">
                                <div className="author-avatar">{testimonial.avatar}</div>
                                <div className="author-info">
                                    <span className="author-name">{testimonial.name}</span>
                                    <span className="author-title">
                                        {testimonial.title}, {testimonial.company}
                                    </span>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>

                <div className="trust-badges">
                    <p>Trusted by companies of all sizes</p>
                    <div className="badges-row">
                        <div className="badge">ISO 27001 Certified</div>
                        <div className="badge">PDPA Compliant</div>
                        <div className="badge">SOC 2 Type II</div>
                        <div className="badge">99.9% Uptime SLA</div>
                    </div>
                </div>
            </div>
        </section>
    );
}
