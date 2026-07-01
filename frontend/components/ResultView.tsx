"use client";

import type { AnalysisResult, RiskLevel } from "@/lib/types";

interface Props {
  result: AnalysisResult;
}

function severityColor(risk: string): string {
  switch (risk.toLowerCase()) {
    case "critical": return "text-red-400 bg-red-950/30 border-red-900";
    case "high":     return "text-orange-400 bg-orange-950/30 border-orange-900";
    case "medium":   return "text-yellow-400 bg-yellow-950/30 border-yellow-900";
    default:         return "text-green-400 bg-green-950/30 border-green-900";
  }
}

function ConfidenceBar({ value }: { value: number }) {
  const pct = Math.min(100, Math.max(0, value));
  const color =
    pct >= 80 ? "#22c55e" :
    pct >= 60 ? "#f59e0b" : "#ef4444";
  return (
    <div className="space-y-1.5">
      <div className="flex justify-between text-xs text-slate-400">
        <span>Confidence</span>
        <span style={{ color }} className="font-mono font-semibold">{pct.toFixed(1)}%</span>
      </div>
      <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
        <div
          className="h-full rounded-full transition-all duration-1000"
          style={{ width: `${pct}%`, background: color }}
        />
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">{title}</p>
      {children}
    </div>
  );
}

function Card({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`border border-slate-800 bg-slate-900/50 rounded-lg p-4 ${className}`}>
      {children}
    </div>
  );
}

function CodeBlock({ code }: { code: string }) {
  return (
    <pre className="bg-slate-950 border border-slate-800 rounded-md px-4 py-3 text-xs font-mono text-lime-400 overflow-x-auto whitespace-pre-wrap break-all">
      {code}
    </pre>
  );
}

export default function ResultView({ result }: Props) {
  const { review } = result;
  const approvedColor = review.approved
    ? "text-green-400 bg-green-950/30 border-green-900"
    : "text-red-400 bg-red-950/30 border-red-900";

  return (
    <div className="space-y-6 fade-in">
      {/* Root Cause */}
      <Section title="Root Cause">
        <Card className="border-cyan-900/60 bg-cyan-950/10">
          <p className="text-white font-medium leading-relaxed">{result.rootCause}</p>
          <div className="mt-4 space-y-3">
            <ConfidenceBar value={result.confidence} />
            <div className="flex gap-2 flex-wrap">
              <span className={`text-xs font-semibold px-2.5 py-1 rounded-full border ${severityColor(result.risk)}`}>
                Risk: {result.risk}
              </span>
              {result.similarIncidents > 0 && (
                <span className="text-xs font-semibold px-2.5 py-1 rounded-full border text-violet-400 bg-violet-950/30 border-violet-900">
                  {result.similarIncidents} similar past incident{result.similarIncidents !== 1 ? "s" : ""}
                </span>
              )}
            </div>
          </div>
        </Card>
      </Section>

      {/* Recommended Action */}
      <Section title="Recommended Action">
        <Card>
          <p className="text-slate-200 text-sm leading-relaxed whitespace-pre-line">{result.recommendedAction}</p>
        </Card>
      </Section>

      {/* Investigation */}
      <Section title="Investigation">
        <Card className="space-y-4">
          <div>
            <p className="text-xs text-slate-500 mb-1.5">Summary</p>
            <p className="text-slate-200 text-sm leading-relaxed">{result.investigation.summary}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500 mb-1.5">Key Findings</p>
            <p className="text-slate-200 text-sm leading-relaxed whitespace-pre-line">{result.investigation.keyFindings}</p>
          </div>
          {result.investigation.dataGaps && (
            <div>
              <p className="text-xs text-slate-500 mb-1.5">Data Gaps</p>
              <p className="text-slate-400 text-sm leading-relaxed">{result.investigation.dataGaps}</p>
            </div>
          )}
        </Card>
      </Section>

      {/* Execution Plan */}
      {result.executionPlan.steps?.length > 0 && (
        <Section title="Execution Plan">
          <div className="space-y-3">
            {result.executionPlan.requires && (
              <p className="text-xs text-slate-500 font-mono">
                Requires: {result.executionPlan.requires}
              </p>
            )}
            {result.executionPlan.steps.map((step) => (
              <Card key={step.order} className="space-y-2.5">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-slate-500 font-mono w-5">
                      {step.order}.
                    </span>
                    <span className="text-xs font-mono uppercase text-slate-400 border border-slate-700 px-1.5 py-0.5 rounded">
                      {step.type}
                    </span>
                    <span className={`text-xs font-semibold px-2 py-0.5 rounded border ${severityColor(step.risk)}`}>
                      {step.risk}
                    </span>
                  </div>
                </div>
                <p className="text-sm text-slate-300 ml-7">{step.purpose}</p>
                <CodeBlock code={`$ ${step.command}`} />
                {step.rollback && (
                  <div className="flex items-start gap-2 ml-7">
                    <span className="text-slate-600 text-xs mt-0.5">↩</span>
                    <p className="text-xs text-slate-500 font-mono">{step.rollback}</p>
                  </div>
                )}
              </Card>
            ))}
          </div>
        </Section>
      )}

      {/* Safety Review */}
      <Section title="Safety Review">
        <Card>
          <div className="flex items-center gap-3 mb-3">
            <span className={`inline-flex items-center gap-1.5 text-sm font-bold px-3 py-1.5 rounded-full border ${approvedColor}`}>
              {review.approved ? (
                <>
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  Approved
                </>
              ) : (
                <>
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  Rejected
                </>
              )}
            </span>
            <span className={`text-xs font-semibold px-2 py-0.5 rounded border ${severityColor(review.overallRisk)}`}>
              {review.overallRisk} risk
            </span>
          </div>
          <p className="text-slate-200 text-sm leading-relaxed">{review.finalVerdict}</p>
          {review.concerns && review.concerns !== "None." && (
            <div className="mt-3 pt-3 border-t border-slate-800">
              <p className="text-xs text-slate-500 mb-1">Concerns</p>
              <p className="text-slate-400 text-sm leading-relaxed">{review.concerns}</p>
            </div>
          )}
          {review.modifications && review.modifications !== "None required." && (
            <div className="mt-3 pt-3 border-t border-slate-800">
              <p className="text-xs text-slate-500 mb-1">Modifications</p>
              <p className="text-slate-400 text-sm leading-relaxed">{review.modifications}</p>
            </div>
          )}
        </Card>
      </Section>
    </div>
  );
}
