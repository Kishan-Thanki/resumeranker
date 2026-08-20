APP_STYLES = """
<style>
    .block-container {
        max-width: 1400px;
        padding-top: 1.25rem;
        padding-bottom: 2.5rem;
    }

    [data-testid="stFileUploader"] button[aria-label="Add files"] {
        display: none;
    }

    .hero {
        padding: 1.5rem 1.6rem;
        border-radius: 22px;
        border: 1px solid rgba(120, 120, 120, 0.20);
        background:
            linear-gradient(135deg, rgba(24, 24, 24, 0.02), rgba(24, 24, 24, 0.08)),
            radial-gradient(circle at top right, rgba(24, 24, 24, 0.08), transparent 34%);
        box-shadow: 0 24px 80px rgba(0, 0, 0, 0.06);
        margin-bottom: 1.25rem;
    }

    .hero-kicker {
        text-transform: uppercase;
        letter-spacing: 0.18em;
        font-size: 0.72rem;
        opacity: 0.72;
        margin-bottom: 0.5rem;
    }

    .hero-title {
        font-size: clamp(2rem, 4vw, 3.2rem);
        line-height: 1.05;
        font-weight: 800;
        margin: 0;
    }

    .hero-subtitle {
        margin: 0.85rem 0 0;
        max-width: 70ch;
        font-size: 1rem;
        line-height: 1.65;
        opacity: 0.8;
    }

    .pill-row {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        margin-top: 1rem;
    }

    .pill {
        display: inline-flex;
        align-items: center;
        gap: 0.35rem;
        padding: 0.35rem 0.7rem;
        border-radius: 999px;
        border: 1px solid rgba(120, 120, 120, 0.22);
        background: rgba(120, 120, 120, 0.08);
        font-size: 0.82rem;
        line-height: 1;
    }

    .stat-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
        gap: 0.75rem;
    }

    .stat-card {
        padding: 0.95rem 1rem;
        border-radius: 18px;
        border: 1px solid rgba(120, 120, 120, 0.22);
        background: rgba(120, 120, 120, 0.05);
        min-height: 102px;
    }

    .stat-label {
        font-size: 0.8rem;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        opacity: 0.68;
        margin-bottom: 0.45rem;
    }

    .stat-value {
        font-size: 1.6rem;
        font-weight: 750;
        line-height: 1.1;
    }

    .stat-note {
        margin-top: 0.35rem;
        font-size: 0.85rem;
        opacity: 0.72;
    }

    .overall-score-panel {
        display: flex;
        align-items: center;
        gap: 1.35rem;
        padding: 1.25rem 1.35rem;
        margin-bottom: 1.25rem;
        border: 1px solid rgba(120, 120, 120, 0.22);
        border-radius: 18px;
        background: rgba(120, 120, 120, 0.05);
    }

    .score-ring {
        --score-color: #0f5132;
        width: 116px;
        aspect-ratio: 1;
        display: grid;
        place-items: center;
        flex: 0 0 auto;
        border-radius: 50%;
        background: conic-gradient(var(--score-color) var(--score), rgba(120, 120, 120, 0.18) 0);
    }

    .score-ring-inner {
        width: 84px;
        aspect-ratio: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        border-radius: 50%;
        background: var(--background-color, #ffffff);
    }

    .score-strong { --score-color: #198754; }
    .score-partial { --score-color: #d39e00; }
    .score-weak { --score-color: #087990; }
    .score-none { --score-color: #dc3545; }

    .score-value {
        font-size: 2rem;
        font-weight: 800;
        line-height: 1;
    }

    .score-unit {
        margin-top: 0.2rem;
        font-size: 0.74rem;
        opacity: 0.65;
    }

    .overall-score-copy h2 {
        margin: 0.18rem 0 0.35rem;
        font-size: 1.45rem;
    }

    .overall-score-copy p {
        margin: 0;
        opacity: 0.72;
    }

    .eyebrow,
    .panel-label {
        font-size: 0.74rem;
        font-weight: 700;
        letter-spacing: 0.1em;
        text-transform: uppercase;
        opacity: 0.68;
    }

    .skill-panel {
        min-height: 84px;
        padding: 0.9rem 1rem;
        border: 1px solid rgba(120, 120, 120, 0.18);
        border-radius: 14px;
        background: rgba(120, 120, 120, 0.045);
    }

    .skill-chip {
        display: inline-block;
        margin: 0.65rem 0.4rem 0 0;
        padding: 0.32rem 0.58rem;
        border: 1px solid transparent;
        border-radius: 999px;
        font-size: 0.82rem;
    }

    .skill-chip-positive {
        color: #0f5132;
        background: #d1e7dd;
        border-color: #badbcc;
    }

    .skill-chip-negative {
        color: #842029;
        background: #f8d7da;
        border-color: #f5c2c7;
    }

    .muted-text {
        display: inline-block;
        margin-top: 0.65rem;
        opacity: 0.68;
    }

    .badge {
        display: inline-flex;
        align-items: center;
        border-radius: 999px;
        padding: 0.22rem 0.6rem;
        font-size: 0.76rem;
        font-weight: 650;
        line-height: 1;
        border: 1px solid transparent;
    }

    .badge-strong {
        color: #0f5132;
        background: #d1e7dd;
        border-color: #badbcc;
    }

    .badge-partial {
        color: #664d03;
        background: #fff3cd;
        border-color: #ffecb5;
    }

    .badge-weak {
        color: #055160;
        background: #cff4fc;
        border-color: #b6effb;
    }

    .badge-none {
        color: #842029;
        background: #f8d7da;
        border-color: #f5c2c7;
    }

    .evidence-item {
        padding: 0.7rem 0.75rem;
        border-radius: 12px;
        background: rgba(255, 255, 255, 0.58);
        border: 1px solid rgba(120, 120, 120, 0.14);
        margin-bottom: 0.65rem;
    }

    .evidence-location {
        font-size: 0.78rem;
        opacity: 0.72;
        margin-bottom: 0.35rem;
    }

    .evidence-text {
        margin: 0;
        line-height: 1.55;
    }
</style>
"""
