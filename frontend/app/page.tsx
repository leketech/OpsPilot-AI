import Link from "next/link";

const AGENTS = [
  {
    name: "Investigation",
    desc: "Queries Kubernetes, Prometheus, GitHub, and ArgoCD for evidence.",
    icon: "⬡",
    color: "text-cyan-400",
  },
  {
    name: "Memory",
    desc: "Surfaces similar past incidents and their resolutions from Redis.",
    icon: "⬡",
    color: "text-violet-400",
  },
  {
    name: "Planner",
    desc: "Reasons over all evidence and generates a structured recommendation.",
    icon: "⬡",
    color: "text-amber-400",
  },
  {
    name: "Reviewer",
    desc: "Validates the plan for safety and requests revision if needed.",
    icon: "⬡",
    color: "text-green-400",
  },
  {
    name: "Executor",
    desc: "Produces a concrete, runnable remediation runbook with rollbacks.",
    icon: "⬡",
    color: "text-rose-400",
  },
];

export default function Home() {
  return (
    <div className="max-w-5xl mx-auto px-6 py-16 space-y-24">
      {/* Hero */}
      <section className="space-y-6">
        <p className="text-xs font-semibold uppercase tracking-[0.25em] text-cyan-400">
          Agentic Incident Response
        </p>
        <h1 className="text-5xl sm:text-6xl font-bold text-white leading-tight">
          From alert to remediation
          <br />
          <span className="text-cyan-400">in minutes, not hours.</span>
        </h1>
        <p className="max-w-2xl text-lg text-slate-400 leading-relaxed">
          OpsPilot AI orchestrates five specialised agents — investigation,
          memory, planning, review, and execution — to autonomously diagnose
          production incidents and propose safe, human-approved remediations.
        </p>
        <div className="flex items-center gap-4 pt-2">
          <Link
            href="/analyze"
            className="inline-flex items-center gap-2 bg-cyan-500 hover:bg-cyan-400 text-black font-semibold px-6 py-3 rounded-lg transition-colors"
          >
            Analyze an Incident
            <span aria-hidden>→</span>
          </Link>
          <a
            href="https://github.com"
            target="_blank"
            rel="noopener noreferrer"
            className="text-slate-400 hover:text-white transition-colors text-sm"
          >
            View source →
          </a>
        </div>
      </section>

      {/* Agent Pipeline */}
      <section className="space-y-8">
        <div>
          <h2 className="text-2xl font-bold text-white">The Agent Pipeline</h2>
          <p className="mt-2 text-slate-400">
            Every incident flows through five coordinated agents. No prompt
            engineering needed — the orchestrator handles it all.
          </p>
        </div>
        <div className="relative">
          {/* Connector line */}
          <div className="absolute left-5 top-8 bottom-8 w-px bg-slate-800" />
          <div className="space-y-4">
            {AGENTS.map((a, i) => (
              <div key={a.name} className="flex gap-4 items-start">
                <div
                  className={`relative z-10 w-10 h-10 rounded-full border border-slate-700 bg-slate-900 flex items-center justify-center shrink-0 ${a.color} font-bold text-sm`}
                >
                  {i + 1}
                </div>
                <div className="border border-slate-800 bg-slate-900/50 rounded-lg p-4 flex-1">
                  <p className={`font-semibold ${a.color}`}>{a.name} Agent</p>
                  <p className="text-sm text-slate-400 mt-0.5">{a.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Feature Grid */}
      <section className="space-y-8">
        <h2 className="text-2xl font-bold text-white">Built for real SRE work</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[
            { title: "Tool-calling ReAct loop", body: "Qwen decides which tools to call. Data comes from real integrations, not hardcoded prompts." },
            { title: "Structured memory", body: "Past incidents — root causes, fixes, reviewer verdicts — are indexed and retrieved on demand." },
            { title: "Safety revision loop", body: "Rejected plans are automatically revised against reviewer concerns, up to two times." },
            { title: "Confidence scoring", body: "Every recommendation includes a numeric confidence score with step-by-step reasoning." },
            { title: "Human approval gate", body: "The reviewer agent validates every plan before it reaches the executor." },
            { title: "Swap-ready integrations", body: "Mock tools are drop-in replaceable with client-go, the Prometheus HTTP API, and GitHub REST." },
          ].map((f) => (
            <div key={f.title} className="border border-slate-800 rounded-lg p-5 bg-slate-900/30">
              <p className="font-semibold text-white text-sm">{f.title}</p>
              <p className="text-slate-400 text-sm mt-1.5 leading-relaxed">{f.body}</p>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
