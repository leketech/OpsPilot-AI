"use client";

import { useState } from "react";
import { SAMPLE_INCIDENT } from "@/lib/api";
import type { IncidentInput, Severity } from "@/lib/types";

interface Props {
  onSubmit: (incident: IncidentInput) => void;
  disabled: boolean;
}

const SEVERITY_OPTIONS: { value: Severity; label: string; color: string }[] = [
  { value: "critical", label: "Critical", color: "text-red-400" },
  { value: "high",     label: "High",     color: "text-orange-400" },
  { value: "medium",   label: "Medium",   color: "text-yellow-400" },
  { value: "low",      label: "Low",      color: "text-blue-400" },
];

const EMPTY: IncidentInput = {
  id: "",
  title: "",
  service: "",
  severity: "high",
  namespace: "",
  cluster: "",
  description: "",
  evidence: { logs: "", metrics: "", events: "" },
};

export default function IncidentForm({ onSubmit, disabled }: Props) {
  const [form, setForm] = useState<IncidentInput>(EMPTY);
  const [showEvidence, setShowEvidence] = useState(false);

  const set = (field: keyof IncidentInput, value: string) =>
    setForm((f) => ({ ...f, [field]: value }));

  const setEvidence = (field: "logs" | "metrics" | "events", value: string) =>
    setForm((f) => ({ ...f, evidence: { ...f.evidence, [field]: value } }));

  function loadSample() {
    setForm(SAMPLE_INCIDENT);
    setShowEvidence(true);
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.title || !form.service || !form.description) return;
    onSubmit({
      ...form,
      id: form.id || `INC-${Date.now()}`,
    });
  }

  const inputCls =
    "w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 transition disabled:opacity-40";
  const labelCls = "block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5";

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="font-bold text-white text-lg">New Incident</h2>
        <button
          type="button"
          onClick={loadSample}
          disabled={disabled}
          className="text-xs text-cyan-400 hover:text-cyan-300 border border-cyan-800 hover:border-cyan-600 px-3 py-1.5 rounded-lg transition disabled:opacity-40"
        >
          Load Sample
        </button>
      </div>

      {/* Title */}
      <div>
        <label className={labelCls}>Title *</label>
        <input
          type="text"
          placeholder="e.g. High CPU — payments-api"
          value={form.title}
          onChange={(e) => set("title", e.target.value)}
          disabled={disabled}
          required
          className={inputCls}
        />
      </div>

      {/* Service + Severity */}
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className={labelCls}>Service *</label>
          <input
            type="text"
            placeholder="payments-api"
            value={form.service}
            onChange={(e) => set("service", e.target.value)}
            disabled={disabled}
            required
            className={inputCls}
          />
        </div>
        <div>
          <label className={labelCls}>Severity</label>
          <select
            value={form.severity}
            onChange={(e) => set("severity", e.target.value as Severity)}
            disabled={disabled}
            className={inputCls}
          >
            {SEVERITY_OPTIONS.map((s) => (
              <option key={s.value} value={s.value}>
                {s.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Namespace + Cluster */}
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className={labelCls}>Namespace</label>
          <input
            type="text"
            placeholder="payments"
            value={form.namespace}
            onChange={(e) => set("namespace", e.target.value)}
            disabled={disabled}
            className={inputCls}
          />
        </div>
        <div>
          <label className={labelCls}>Cluster</label>
          <input
            type="text"
            placeholder="prod-us-east-1"
            value={form.cluster}
            onChange={(e) => set("cluster", e.target.value)}
            disabled={disabled}
            className={inputCls}
          />
        </div>
      </div>

      {/* Description */}
      <div>
        <label className={labelCls}>Description *</label>
        <textarea
          placeholder="Describe what is happening and when it started…"
          value={form.description}
          onChange={(e) => set("description", e.target.value)}
          disabled={disabled}
          required
          rows={4}
          className={`${inputCls} resize-none`}
        />
      </div>

      {/* Evidence (collapsible) */}
      <div className="border border-slate-800 rounded-lg overflow-hidden">
        <button
          type="button"
          onClick={() => setShowEvidence((v) => !v)}
          className="w-full flex items-center justify-between px-4 py-3 text-sm text-slate-300 hover:bg-slate-800/40 transition"
        >
          <span className="font-medium">Evidence (optional)</span>
          <span className="text-slate-500 text-xs">{showEvidence ? "▲" : "▼"}</span>
        </button>

        {showEvidence && (
          <div className="border-t border-slate-800 p-4 space-y-4">
            <div>
              <label className={labelCls}>Logs</label>
              <textarea
                placeholder="Paste relevant log lines…"
                value={form.evidence?.logs ?? ""}
                onChange={(e) => setEvidence("logs", e.target.value)}
                disabled={disabled}
                rows={5}
                className={`${inputCls} resize-none font-mono text-xs`}
              />
            </div>
            <div>
              <label className={labelCls}>Metrics</label>
              <textarea
                placeholder="CPU %, memory, error rate, latency…"
                value={form.evidence?.metrics ?? ""}
                onChange={(e) => setEvidence("metrics", e.target.value)}
                disabled={disabled}
                rows={4}
                className={`${inputCls} resize-none font-mono text-xs`}
              />
            </div>
            <div>
              <label className={labelCls}>Events</label>
              <textarea
                placeholder="kubectl events, PagerDuty alerts…"
                value={form.evidence?.events ?? ""}
                onChange={(e) => setEvidence("events", e.target.value)}
                disabled={disabled}
                rows={4}
                className={`${inputCls} resize-none font-mono text-xs`}
              />
            </div>
          </div>
        )}
      </div>

      {/* Submit */}
      <button
        type="submit"
        disabled={disabled || !form.title || !form.service || !form.description}
        className="w-full bg-cyan-500 hover:bg-cyan-400 disabled:bg-slate-700 disabled:text-slate-500 text-black font-semibold py-3 rounded-lg transition text-sm"
      >
        {disabled ? "Analyzing…" : "Analyze Incident"}
      </button>
    </form>
  );
}
