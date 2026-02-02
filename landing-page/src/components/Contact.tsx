import { useState, FormEvent } from 'react';

export default function Contact() {
    const [formData, setFormData] = useState({
        name: '',
        email: '',
        company: '',
        message: '',
    });
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isSubmitted, setIsSubmitted] = useState(false);

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setIsSubmitting(true);

        // Simulate form submission
        await new Promise((resolve) => setTimeout(resolve, 1000));

        setIsSubmitting(false);
        setIsSubmitted(true);
        setFormData({ name: '', email: '', company: '', message: '' });
    };

    return (
        <section id="contact" className="contact">
            <div className="container">
                <div className="contact-content">
                    <div className="contact-info">
                        <span className="section-badge">Contact Us</span>
                        <h2 className="section-title">
                            Let's <span className="gradient-text">talk</span>
                        </h2>
                        <p className="section-description">
                            Have questions about our CRM? Want to see a demo? We'd love to hear
                            from you. Fill out the form and our team will get back to you within
                            24 hours.
                        </p>

                        <div className="contact-methods">
                            <div className="contact-method">
                                <div className="method-icon">📧</div>
                                <div className="method-details">
                                    <span className="method-label">Email us</span>
                                    <a href="mailto:hello@kilangbatik.com">hello@kilangbatik.com</a>
                                </div>
                            </div>
                            <div className="contact-method">
                                <div className="method-icon">📞</div>
                                <div className="method-details">
                                    <span className="method-label">Call us</span>
                                    <a href="tel:+60123456789">+60 12-345 6789</a>
                                </div>
                            </div>
                            <div className="contact-method">
                                <div className="method-icon">📍</div>
                                <div className="method-details">
                                    <span className="method-label">Visit us</span>
                                    <span>Kuala Lumpur, Malaysia</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="contact-form-wrapper">
                        {isSubmitted ? (
                            <div className="success-message">
                                <div className="success-icon">✓</div>
                                <h3>Thank you!</h3>
                                <p>
                                    We've received your message and will get back to you within 24
                                    hours.
                                </p>
                                <button
                                    className="btn btn-outline"
                                    onClick={() => setIsSubmitted(false)}
                                >
                                    Send another message
                                </button>
                            </div>
                        ) : (
                            <form className="contact-form" onSubmit={handleSubmit}>
                                <div className="form-group">
                                    <label htmlFor="name">Full Name</label>
                                    <input
                                        type="text"
                                        id="name"
                                        value={formData.name}
                                        onChange={(e) =>
                                            setFormData({ ...formData, name: e.target.value })
                                        }
                                        placeholder="John Doe"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor="email">Email Address</label>
                                    <input
                                        type="email"
                                        id="email"
                                        value={formData.email}
                                        onChange={(e) =>
                                            setFormData({ ...formData, email: e.target.value })
                                        }
                                        placeholder="john@company.com"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor="company">Company Name</label>
                                    <input
                                        type="text"
                                        id="company"
                                        value={formData.company}
                                        onChange={(e) =>
                                            setFormData({ ...formData, company: e.target.value })
                                        }
                                        placeholder="Your Company Sdn Bhd"
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor="message">Message</label>
                                    <textarea
                                        id="message"
                                        value={formData.message}
                                        onChange={(e) =>
                                            setFormData({ ...formData, message: e.target.value })
                                        }
                                        placeholder="Tell us about your needs..."
                                        rows={4}
                                        required
                                    ></textarea>
                                </div>
                                <button
                                    type="submit"
                                    className="btn btn-primary btn-block"
                                    disabled={isSubmitting}
                                >
                                    {isSubmitting ? 'Sending...' : 'Send Message'}
                                </button>
                            </form>
                        )}
                    </div>
                </div>
            </div>
        </section>
    );
}
