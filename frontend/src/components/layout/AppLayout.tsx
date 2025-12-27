import React from 'react';
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useTheme } from '../../context/ThemeContext';
import { useChat } from '../../context/ChatContext';
import { useLanguage } from '../../context/LanguageContext';
import { 
  MessageSquare, 
  Users, 
  PlusCircle, 
  User, 
  LogOut, 
  Sparkles,
  Menu,
  Sun,
  Moon,
  Trash2,
  Languages,
  ChevronDown,
  ChevronRight,
  Globe,
  LayoutGrid,
  X
} from 'lucide-react';
import { cn } from '../../lib/utils';
import { useState } from 'react';

export const AppLayout: React.FC = () => {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const { language, setLanguage, t } = useLanguage();
  const { sessions, deleteSession } = useChat();
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());

  const toggleGroup = (charId: string) => {
    setCollapsedGroups(prev => {
      const next = new Set(prev);
      if (next.has(charId)) {
        next.delete(charId);
      } else {
        next.add(charId);
      }
      return next;
    });
  };

  const handleDeleteSession = async (e: React.MouseEvent, sessionId: string) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (window.confirm(t('nav.confirmDelete'))) {
      await deleteSession(sessionId);
      // If we are currently on the deleted session page, redirect to dashboard
      if (location.pathname === `/chat/${sessionId}`) {
        navigate('/');
      }
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const navItems = [
    { to: '/', icon: Users, label: t('nav.companions') },
    { to: '/market', icon: Globe, label: t('nav.market') },
    { to: '/chat', icon: MessageSquare, label: t('nav.chats') },
    { to: '/create', icon: PlusCircle, label: t('nav.create') },
    { to: '/profile', icon: User, label: t('nav.profile') },
  ];

  const groupedSessions = React.useMemo(() => {
    const groups: Record<string, typeof sessions> = {};
    sessions.forEach(session => {
      if (!groups[session.characterId]) {
        groups[session.characterId] = [];
      }
      groups[session.characterId].push(session);
    });
    return groups;
  }, [sessions]);

  return (
    <div className="min-h-screen flex font-sans text-slate-900 dark:text-slate-100 transition-colors duration-200">
      {/* Mobile Menu Overlay */}
      {mobileMenuOpen && (
        <div 
          className="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-40 lg:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={cn(
        "fixed lg:static inset-y-0 left-0 z-50 w-64 bg-white/80 dark:bg-slate-900/80 backdrop-blur-xl border-r border-slate-200/50 dark:border-slate-800/50 text-slate-900 dark:text-slate-100 transform transition-transform duration-300 ease-out lg:translate-x-0 flex flex-col shadow-2xl lg:shadow-none",
        mobileMenuOpen ? "translate-x-0" : "-translate-x-full"
      )}>
        <div className="p-6 flex items-center gap-3 border-b border-slate-200/50 dark:border-slate-800/50">
          <div className="relative w-8 h-8 flex items-center justify-center">
            <div className="absolute inset-0 bg-gradient-to-tr from-cyan-500 to-violet-500 rounded-xl blur opacity-60 animate-pulse"></div>
            <div className="relative w-full h-full bg-gradient-to-tr from-cyan-500 to-violet-500 rounded-xl flex items-center justify-center shadow-lg shadow-violet-500/20">
               <Sparkles className="w-5 h-5 text-white" />
            </div>
          </div>
          <span className="font-bold text-xl tracking-tight bg-gradient-to-r from-slate-900 to-slate-600 dark:from-white dark:to-slate-400 bg-clip-text text-transparent">
            AI Companion
          </span>
        </div>

        <nav className="flex-none p-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              onClick={() => setMobileMenuOpen(false)}
              className={({ isActive }) => cn(
                "flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group font-medium",
                isActive 
                  ? "bg-gradient-to-r from-violet-600 to-indigo-600 text-white shadow-lg shadow-violet-500/30 translate-x-1" 
                  : "text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800/50 hover:text-slate-900 dark:hover:text-slate-100"
              )}
            >
              <item.icon className={cn("w-5 h-5 transition-transform group-hover:scale-110", ({ isActive }: any) => isActive ? "animate-pulse" : "")} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        {/* Recent Chats Section */}
        <div className="px-4 pb-4 overflow-y-auto flex-1 min-h-0 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-slate-700 scrollbar-track-transparent">
          <h3 className="text-xs font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider mb-3 px-2 mt-4">{t('nav.recentChats')}</h3>
          <div className="space-y-2">
            {Object.keys(groupedSessions).length === 0 ? (
              <p className="text-xs text-slate-400 px-2 italic">{t('nav.noChats')}</p>
            ) : (
              Object.entries(groupedSessions).map(([charId, charSessions]) => {
                const character = charSessions[0]; // Use first session to get character info
                const isCollapsed = collapsedGroups.has(charId);
                
                return (
                  <div key={charId} className="space-y-1">
                    <button 
                      onClick={() => toggleGroup(charId)}
                      className="w-full flex items-center gap-2 px-2 py-1.5 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100/50 dark:hover:bg-slate-800/30 rounded-lg transition-all group"
                    >
                      {isCollapsed ? <ChevronRight className="w-3.5 h-3.5 opacity-70" /> : <ChevronDown className="w-3.5 h-3.5 opacity-70" />}
                      <img src={character.characterAvatar} className="w-5 h-5 rounded-full object-cover ring-2 ring-white dark:ring-slate-800" alt="" />
                      <span className="text-xs font-bold uppercase tracking-wide opacity-80 flex-1 text-left truncate">{character.characterName}</span>
                      <span className="text-[10px] bg-slate-200 dark:bg-slate-800 px-1.5 py-0.5 rounded-full opacity-60 group-hover:opacity-100 transition-opacity">
                        {charSessions.length}
                      </span>
                    </button>
                    
                    <div className={cn(
                      "space-y-1 overflow-hidden transition-all duration-300 ease-in-out",
                      isCollapsed ? "max-h-0 opacity-0" : "max-h-[500px] opacity-100"
                    )}>
                      {charSessions.map(session => (
                        <NavLink
                          key={session.id}
                          to={`/chat/${session.id}`}
                          onClick={() => setMobileMenuOpen(false)}
                          className={({ isActive }) => cn(
                            "block pl-9 pr-3 py-2 rounded-lg transition-all text-sm group relative border border-transparent ml-2",
                            isActive 
                              ? "bg-violet-50 dark:bg-violet-500/10 text-violet-600 dark:text-violet-300 font-medium border-violet-100 dark:border-violet-500/20" 
                              : "text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800/50 hover:text-slate-900 dark:hover:text-slate-200"
                          )}
                        >
                          <div className="min-w-0 mr-6">
                            <div className="truncate text-[10px] opacity-60 mb-0.5">
                              {new Date(session.updatedAt).toLocaleDateString()}
                            </div>
                            <div className="truncate text-xs">
                              {session.lastMessage}
                            </div>
                          </div>
                          <button
                            onClick={(e) => handleDeleteSession(e, session.id)}
                            className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md opacity-0 group-hover:opacity-100 transition-all focus:opacity-100"
                            title={t('nav.deleteChat')}
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </NavLink>
                      ))}
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        <div className="p-4 border-t border-slate-200/50 dark:border-slate-800/50 space-y-2 bg-white/50 dark:bg-slate-900/50 backdrop-blur-sm">
          <div className="relative group">
            <Languages className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none group-hover:text-violet-500 transition-colors" />
            <select
              value={language}
              onChange={(e) => setLanguage(e.target.value as 'en' | 'zh')}
              className="w-full appearance-none bg-slate-50 dark:bg-slate-800/50 hover:bg-white dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 pl-11 pr-10 py-2.5 rounded-xl transition-all cursor-pointer outline-none border border-transparent hover:border-violet-200 dark:hover:border-violet-800 focus:ring-2 focus:ring-violet-500/20 text-sm font-medium"
            >
              <option value="zh" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">中文 (Chinese)</option>
              <option value="en" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">English</option>
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 pointer-events-none group-hover:text-violet-500 transition-colors" />
          </div>

           <button
            onClick={toggleTheme}
            className="flex items-center gap-3 px-4 py-2.5 w-full text-slate-600 dark:text-slate-400 hover:bg-white dark:hover:bg-slate-800 hover:text-amber-500 dark:hover:text-amber-400 rounded-xl transition-all text-sm font-medium border border-transparent hover:border-amber-200 dark:hover:border-amber-900/50 hover:shadow-sm"
          >
            {theme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
            <span>{theme === 'dark' ? 'Light Mode' : 'Dark Mode'}</span>
          </button>

          <div className="flex items-center gap-3 px-4 py-3 mb-2 bg-gradient-to-r from-cyan-500/5 to-blue-500/5 border border-cyan-100 dark:border-cyan-900/30 rounded-xl">
            <div className="flex-1">
              <p className="text-[10px] text-slate-500 uppercase tracking-wider font-bold">{t('nav.credits')}</p>
              <p className="text-lg font-bold bg-gradient-to-r from-cyan-500 to-blue-500 bg-clip-text text-transparent">{user?.credits || 0}</p>
            </div>
            <button className="text-xs bg-cyan-500 text-white px-3 py-1.5 rounded-lg hover:bg-cyan-600 transition shadow-lg shadow-cyan-500/30 font-medium">
              {t('nav.topUp')}
            </button>
          </div>
          
          <button 
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-2.5 w-full text-slate-500 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/10 rounded-xl transition-colors text-sm font-medium"
          >
            <LogOut className="w-4 h-4" />
            <span>{t('nav.signOut')}</span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Mobile Header */}
        <header className="lg:hidden h-16 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-4 transition-colors">
          <button onClick={() => setMobileMenuOpen(true)} className="p-2 text-slate-600 dark:text-slate-400">
            <Menu className="w-6 h-6" />
          </button>
          <span className="font-bold text-lg text-slate-800 dark:text-slate-100">AI Companion</span>
          <div className="w-8" /> {/* Spacer */}
        </header>

        <div className="flex-1 overflow-auto p-4 lg:p-8">
          <div className="max-w-6xl mx-auto h-full">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
};
