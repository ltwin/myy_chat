import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../context/LanguageContext';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { motion } from 'framer-motion';
import { Sparkles, Globe, MessageSquare, ArrowRight, Zap, Heart, Share2 } from 'lucide-react';
import { cn } from '../lib/utils';

export const LandingPage: React.FC = () => {
  const { t } = useLanguage();
  const { isAuthenticated } = useAuth();
  const { theme } = useTheme();
  const navigate = useNavigate();

  const handleGetStarted = () => {
    if (isAuthenticated) {
      navigate('/app');
    } else {
      navigate('/login');
    }
  };

  const features = [
    {
      icon: Heart,
      title: t('landing.feature.1.title'),
      desc: t('landing.feature.1.desc'),
      color: 'from-pink-500 to-rose-500',
      delay: 0.1
    },
    {
      icon: Globe,
      title: t('landing.feature.2.title'),
      desc: t('landing.feature.2.desc'),
      color: 'from-cyan-500 to-blue-500',
      delay: 0.2
    },
    {
      icon: MessageSquare,
      title: t('landing.feature.3.title'),
      desc: t('landing.feature.3.desc'),
      color: 'from-violet-500 to-purple-500',
      delay: 0.3
    }
  ];

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-[#0f172a] text-slate-900 dark:text-slate-100 overflow-hidden selection:bg-violet-500/30">
      {/* Background Gradients */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[500px] h-[500px] bg-purple-500/20 rounded-full blur-[100px] animate-pulse"></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[500px] h-[500px] bg-cyan-500/20 rounded-full blur-[100px] animate-pulse [animation-delay:2s]"></div>
      </div>

      {/* Navbar */}
      <nav className="relative z-10 px-6 py-6 flex items-center justify-between max-w-7xl mx-auto">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-lg shadow-violet-500/20">
            <Sparkles className="w-5 h-5 text-white" />
          </div>
          <span className="font-bold text-xl tracking-tight">AI Companion</span>
        </div>
        <button 
          onClick={() => navigate('/login')}
          className="px-5 py-2 rounded-full font-medium text-sm transition-colors hover:bg-slate-100 dark:hover:bg-slate-800"
        >
          {isAuthenticated ? 'Dashboard' : 'Sign In'}
        </button>
      </nav>

      {/* Hero Section */}
      <main className="relative z-10 pt-20 pb-32 px-6">
        <div className="max-w-4xl mx-auto text-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-violet-50 dark:bg-violet-900/30 text-violet-600 dark:text-violet-300 text-xs font-semibold mb-6 border border-violet-100 dark:border-violet-700/50">
              <Zap className="w-3.5 h-3.5" />
              Next Gen AI Platform
            </span>
            <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-8 leading-tight">
              {t('landing.hero.title')} <br className="hidden md:block" />
              <span className="bg-gradient-to-r from-violet-600 via-fuchsia-500 to-indigo-600 bg-clip-text text-transparent animate-gradient-x">
                {t('landing.hero.highlight')}
              </span>
            </h1>
            <p className="text-xl md:text-2xl text-slate-600 dark:text-slate-400 mb-10 max-w-2xl mx-auto leading-relaxed">
              {t('landing.hero.subtitle')}
            </p>
            
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <button 
                onClick={handleGetStarted}
                className="w-full sm:w-auto px-8 py-4 rounded-full bg-slate-900 dark:bg-white text-white dark:text-slate-900 font-bold text-lg hover:scale-105 transition-transform shadow-xl shadow-slate-900/20 dark:shadow-white/10 flex items-center justify-center gap-2"
              >
                {t('landing.cta.primary')}
                <ArrowRight className="w-5 h-5" />
              </button>
              <button 
                onClick={() => navigate(isAuthenticated ? '/app/market' : '/login')}
                className="w-full sm:w-auto px-8 py-4 rounded-full bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 font-bold text-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center gap-2"
              >
                {t('landing.cta.secondary')}
              </button>
            </div>
          </motion.div>

          {/* Feature Cards */}
          <div className="grid md:grid-cols-3 gap-6 mt-24 text-left">
            {features.map((feature, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: feature.delay, duration: 0.6 }}
                className="p-6 rounded-3xl bg-white/50 dark:bg-slate-900/50 backdrop-blur-xl border border-white/50 dark:border-slate-700/50 hover:border-violet-500/30 transition-all hover:shadow-2xl hover:shadow-violet-500/10 group"
              >
                <div className={cn(
                  "w-12 h-12 rounded-2xl flex items-center justify-center mb-4 bg-gradient-to-br shadow-lg",
                  feature.color
                )}>
                  <feature.icon className="w-6 h-6 text-white" />
                </div>
                <h3 className="text-xl font-bold mb-2 text-slate-900 dark:text-white group-hover:text-violet-600 dark:group-hover:text-violet-400 transition-colors">
                  {feature.title}
                </h3>
                <p className="text-slate-500 dark:text-slate-400 leading-relaxed">
                  {feature.desc}
                </p>
              </motion.div>
            ))}
          </div>
        </div>
      </main>

      {/* Decorative Footer */}
      <footer className="py-8 text-center text-slate-400 text-sm">
        <p>© 2025 AI Companion Platform. Built with ❤️</p>
      </footer>
    </div>
  );
};
