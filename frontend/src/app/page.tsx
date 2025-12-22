/**
 * Home page component
 *
 * This is the main landing page of the MYY Chat application.
 * It serves as the entry point for users and will display:
 * - Welcome message
 * - Feature highlights
 * - Call-to-action buttons
 * - Navigation to main application sections
 *
 * @returns The home page component
 */
export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="z-10 w-full max-w-5xl items-center justify-center font-mono text-sm">
        <div className="flex flex-col items-center gap-8">
          <h1 className="text-6xl font-bold text-center bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            MYY Chat
          </h1>

          <p className="text-xl text-center text-muted-foreground max-w-2xl">
            Enterprise-grade AI companion platform with intelligent conversation capabilities
          </p>

          <div className="flex gap-4 mt-8">
            <a
              href="/chat"
              className="px-6 py-3 bg-primary text-primary-foreground rounded-lg font-semibold hover:opacity-90 transition-opacity"
            >
              Start Chatting
            </a>
            <a
              href="/companions"
              className="px-6 py-3 bg-secondary text-secondary-foreground rounded-lg font-semibold hover:opacity-90 transition-opacity"
            >
              Browse Companions
            </a>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-16 w-full">
            <FeatureCard
              title="Smart Conversations"
              description="AI-powered conversations with context awareness and memory"
            />
            <FeatureCard
              title="Custom Companions"
              description="Create and customize AI companions for specific tasks"
            />
            <FeatureCard
              title="Secure & Scalable"
              description="Enterprise-grade security with scalable architecture"
            />
          </div>
        </div>
      </div>
    </main>
  );
}

/**
 * Feature card component props
 */
interface FeatureCardProps {
  title: string;
  description: string;
}

/**
 * Feature card component for displaying platform features
 *
 * @param props - Component props
 * @param props.title - Feature title
 * @param props.description - Feature description
 */
function FeatureCard({ title, description }: FeatureCardProps) {
  return (
    <div className="p-6 bg-card rounded-lg border border-border hover:border-primary transition-colors">
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}
