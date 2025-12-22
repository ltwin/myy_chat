# MYY Chat Frontend

Enterprise-grade AI companion platform frontend built with Next.js 14, TypeScript, and Tailwind CSS.

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript 5.x
- **Styling**: Tailwind CSS 3.x
- **UI Components**: Shadcn/ui (Radix UI primitives)
- **State Management**: Zustand 4.x
- **Icons**: Lucide React

## Getting Started

### Prerequisites

- Node.js 18.x or later
- npm, yarn, or pnpm

### Installation

```bash
# Install dependencies
npm install

# Copy environment variables
cp .env.example .env.local

# Run development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to view the application.

## Project Structure

```
frontend/
├── src/
│   ├── app/              # Next.js App Router pages
│   │   ├── layout.tsx    # Root layout
│   │   └── page.tsx      # Home page
│   ├── components/
│   │   └── ui/          # Shadcn/ui components
│   ├── lib/
│   │   ├── api/         # API client and utilities
│   │   ├── hooks/       # Custom React hooks
│   │   └── utils/       # Utility functions
│   └── styles/
│       └── globals.css  # Global styles and Tailwind
├── public/              # Static assets
├── tests/
│   └── e2e/            # End-to-end tests
└── config files
```

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint
- `npm run type-check` - Run TypeScript type checking

## Adding Shadcn/UI Components

```bash
# Example: Add a button component
npx shadcn-ui@latest add button

# Example: Add a dialog component
npx shadcn-ui@latest add dialog
```

## TypeScript Configuration

This project uses strict TypeScript configuration:

- Strict mode enabled
- No implicit any
- No unchecked indexed access
- Unused locals and parameters warnings
- No implicit returns

## Environment Variables

Create a `.env.local` file based on `.env.example`:

- `NEXT_PUBLIC_API_URL` - Backend API URL (default: http://localhost:8080)

## Development Guidelines

1. **Type Safety**: Always define proper TypeScript types and interfaces
2. **Components**: Use functional components with TypeScript
3. **Styling**: Use Tailwind CSS utility classes, extract to components when needed
4. **API Calls**: Centralize API logic in `src/lib/api`
5. **State**: Use Zustand for global state, React hooks for local state
6. **Documentation**: Add JSDoc comments for all exported functions and components

## Building for Production

```bash
# Create optimized production build
npm run build

# Start production server
npm run start
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Shadcn/ui Documentation](https://ui.shadcn.com)
- [Zustand Documentation](https://docs.pmnd.rs/zustand)
