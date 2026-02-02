import { useState } from 'react';

const faqs = [
    {
        question: 'How long is the free trial?',
        answer:
            'Our free trial lasts 14 days with full access to all features of the Professional plan. No credit card required to start.',
    },
    {
        question: 'Can I import data from my current CRM?',
        answer:
            'Yes! We support importing data from Salesforce, HubSpot, Zoho, and CSV files. Our team can also assist with custom migrations for Enterprise customers.',
    },
    {
        question: 'Is my data secure?',
        answer:
            'Absolutely. We use bank-grade encryption (AES-256), have SOC 2 Type II certification, and comply with PDPA regulations. Your data is backed up daily and stored in secure Malaysian data centers.',
    },
    {
        question: 'Do you offer training and support?',
        answer:
            'Yes! All plans include email support and access to our knowledge base. Professional and Enterprise plans include priority support, and Enterprise customers get dedicated onboarding and training.',
    },
    {
        question: 'Can I change plans later?',
        answer:
            "Of course! You can upgrade or downgrade your plan at any time. If you upgrade, you'll be prorated for the remainder of your billing cycle.",
    },
    {
        question: 'Do you offer discounts for non-profits or startups?',
        answer:
            'Yes, we offer special pricing for registered non-profits and early-stage startups. Contact our sales team with documentation for consideration.',
    },
    {
        question: 'What payment methods do you accept?',
        answer:
            'We accept all major credit cards (Visa, Mastercard, AMEX), FPX online banking, and bank transfers for annual Enterprise contracts.',
    },
    {
        question: 'Is there a contract or commitment?',
        answer:
            'Monthly plans are pay-as-you-go with no long-term commitment. You can cancel anytime. Annual plans offer a 20% discount and are billed upfront.',
    },
];

export default function FAQ() {
    const [openIndex, setOpenIndex] = useState<number | null>(0);

    return (
        <section id="faq" className="faq">
            <div className="container">
                <div className="section-header">
                    <span className="section-badge">FAQ</span>
                    <h2 className="section-title">
                        Frequently <span className="gradient-text">asked questions</span>
                    </h2>
                    <p className="section-description">
                        Got questions? We've got answers. Can't find what you're looking for?{' '}
                        <a href="#contact">Contact our team</a>.
                    </p>
                </div>

                <div className="faq-list">
                    {faqs.map((faq, index) => (
                        <div
                            key={index}
                            className={`faq-item ${openIndex === index ? 'open' : ''}`}
                        >
                            <button
                                className="faq-question"
                                onClick={() => setOpenIndex(openIndex === index ? null : index)}
                                aria-expanded={openIndex === index}
                            >
                                <span>{faq.question}</span>
                                <svg
                                    className="faq-icon"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                >
                                    <path d="M19 9l-7 7-7-7" />
                                </svg>
                            </button>
                            <div className="faq-answer">
                                <p>{faq.answer}</p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}
