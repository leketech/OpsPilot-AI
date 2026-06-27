export default function Home() {
  return (
    <main className="min-h-screen px-6 py-10">
      <section className="mx-auto flex min-h-[calc(100vh-5rem)] max-w-5xl flex-col justify-center">
        <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">
          OpsPilot-AI
        </p>
        <h1 className="mt-4 max-w-3xl text-4xl font-bold tracking-normal text-white sm:text-6xl">
          Incident response intelligence for modern DevOps teams.
        </h1>
        <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-300">
          Investigate production issues, retrieve operational memory, analyze
          signals with Qwen, and prepare safe remediation workflows with human
          approval.
        </p>
      </section>
    </main>
  );
}
