import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import type { ChatSession, Message } from '../types';
import { mockApi } from '../services/mockApi';
import { Send, User as UserIcon, Loader2, Image as ImageIcon, Sparkles } from 'lucide-react';
import { cn } from '../lib/utils';
import { useAuth } from '../context/AuthContext';
import { useChat } from '../context/ChatContext';
import { useLanguage } from '../context/LanguageContext';

export const ChatPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { refreshSessions } = useChat();
  const { t } = useLanguage();
  const [session, setSession] = useState<ChatSession | null>(null);
  const [loading, setLoading] = useState(true);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!sessionId) {
      // If no session ID, maybe show a "Select a chat" or redirect
      // For now, let's just load the most recent session or show empty state
      loadRecentSession();
    } else {
      loadSession(sessionId);
    }
  }, [sessionId]);

  const loadRecentSession = async () => {
    const sessions = await mockApi.chat.getSessions();
    if (sessions.length > 0) {
      navigate(`/chat/${sessions[0].id}`, { replace: true });
    } else {
      setLoading(false); // No sessions
    }
  };

  const loadSession = async (id: string) => {
    setLoading(true);
    try {
      const data = await mockApi.chat.getSession(id);
      if (data) {
        setSession(data);
      } else {
        navigate('/'); // Invalid session
      }
    } finally {
      setLoading(false);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [session?.messages, sending]);

  const handleSend = async (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || sending || !session) return;

    const content = input.trim();
    setInput('');
    setSending(true);

    try {
      // Optimistic update
      const tempMsg: Message = {
        id: `temp_${Date.now()}`,
        role: 'user',
        content,
        timestamp: Date.now()
      };
      
      setSession(prev => prev ? {
        ...prev,
        messages: [...prev.messages, tempMsg]
      } : null);

      await mockApi.chat.sendMessage(session.id, content);
      
      // Update sidebar list with new last message
      refreshSessions();

      // Reload to get AI response
      const updatedSession = await mockApi.chat.getSession(session.id);
      if (updatedSession) {
        setSession(updatedSession);
      }
    } catch (error) {
      console.error('Failed to send', error);
      // Revert optimistic update ideally, but for mock we skip
    } finally {
      setSending(false);
    }
  };

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    );
  }

  if (!sessionId && !loading && !session) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-center p-8">
        <div className="w-24 h-24 bg-gradient-to-tr from-violet-100 to-indigo-100 dark:from-slate-800 dark:to-slate-800 rounded-full flex items-center justify-center mb-6 shadow-xl shadow-violet-500/10 animate-float">
          <MessageSquare className="w-10 h-10 text-violet-500 dark:text-violet-400" />
        </div>
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white">{t('chat.noActive')}</h2>
        <p className="text-slate-500 dark:text-slate-400 mt-2 mb-8 max-w-sm mx-auto">{t('chat.startPrompt')}</p>
        <button 
          onClick={() => navigate('/')}
          className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white px-8 py-3.5 rounded-xl transition-all shadow-lg shadow-violet-500/30 hover:shadow-violet-500/50 hover:-translate-y-0.5 font-medium"
        >
          {t('chat.exploreBtn')}
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-[calc(100vh-theme(spacing.32))] bg-white/70 dark:bg-slate-900/60 backdrop-blur-xl rounded-2xl shadow-2xl border border-white/50 dark:border-slate-700/50 overflow-hidden transition-all">
      {/* Chat Header */}
      <div className="h-20 border-b border-slate-200/50 dark:border-slate-800/50 px-6 flex items-center justify-between bg-white/50 dark:bg-slate-900/50 backdrop-blur-md sticky top-0 z-10">
        <div className="flex items-center gap-4">
          <div className="relative">
             <div className="absolute inset-0 bg-gradient-to-tr from-violet-500 to-fuchsia-500 rounded-full blur opacity-40"></div>
             <img 
               src={session?.characterAvatar} 
               alt={session?.characterName} 
               className="relative w-12 h-12 rounded-full object-cover border-2 border-white dark:border-slate-700 shadow-md"
             />
             <span className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white dark:border-slate-900 rounded-full"></span>
          </div>
          <div>
            <h2 className="font-bold text-lg text-slate-900 dark:text-white">{session?.characterName}</h2>
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-violet-600 dark:text-violet-400 font-medium bg-violet-100 dark:bg-violet-900/30 px-2 py-0.5 rounded-full">{t('chat.online')}</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button className="p-2.5 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-xl text-slate-400 hover:text-violet-500 transition-colors">
            <ImageIcon className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-6 space-y-8 bg-transparent">
        {session?.messages.map((msg, idx) => {
          const isUser = msg.role === 'user';
          return (
            <div 
              key={msg.id} 
              className={cn(
                "flex gap-4 max-w-3xl animate-in fade-in slide-in-from-bottom-4 duration-500",
                isUser ? "ml-auto flex-row-reverse" : ""
              )}
            >
              <div className={cn(
                "w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center shadow-md border-2 border-white dark:border-slate-700",
                isUser ? "bg-slate-200 dark:bg-slate-700" : "bg-white dark:bg-slate-800"
              )}>
                {isUser ? (
                  <UserIcon className="w-5 h-5 text-slate-500 dark:text-slate-300" />
                ) : (
                  <img src={session?.characterAvatar} className="w-10 h-10 rounded-full object-cover" />
                )}
              </div>
              
              <div className={cn(
                "group relative px-6 py-4 rounded-2xl text-[15px] leading-relaxed shadow-sm max-w-[85%]",
                isUser 
                  ? "bg-gradient-to-br from-violet-600 to-indigo-600 text-white rounded-tr-sm shadow-violet-500/20" 
                  : "bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 border border-slate-100 dark:border-slate-700 rounded-tl-sm shadow-sm"
              )}>
                {msg.content}
                <div className={cn(
                  "text-[10px] mt-2 font-medium opacity-60",
                  isUser ? "text-violet-100 text-right" : "text-slate-400"
                )}>
                  {new Date(msg.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                </div>
              </div>
            </div>
          );
        })}
        
        {sending && (
          <div className="flex gap-4 max-w-3xl animate-in fade-in">
            <img src={session?.characterAvatar} className="w-10 h-10 rounded-full object-cover border-2 border-white dark:border-slate-700 shadow-md" />
            <div className="bg-white dark:bg-slate-800 px-5 py-4 rounded-2xl rounded-tl-sm border border-slate-100 dark:border-slate-700 shadow-sm flex items-center gap-1.5">
              <span className="w-2 h-2 bg-violet-400 rounded-full animate-bounce [animation-delay:-0.3s]"></span>
              <span className="w-2 h-2 bg-violet-400 rounded-full animate-bounce [animation-delay:-0.15s]"></span>
              <span className="w-2 h-2 bg-violet-400 rounded-full animate-bounce"></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="p-6 bg-white/50 dark:bg-slate-900/50 backdrop-blur-md border-t border-slate-200/50 dark:border-slate-800/50">
        <form onSubmit={handleSend} className="relative flex items-end gap-3 max-w-4xl mx-auto">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder={t('chat.typePlaceholder')}
            className="flex-1 bg-white dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white rounded-2xl px-5 py-4 focus:ring-2 focus:ring-violet-500/20 focus:border-violet-500 focus:bg-white dark:focus:bg-slate-800 transition-all placeholder:text-slate-400 shadow-sm"
            disabled={sending}
          />
          <button 
            type="submit"
            disabled={!input.trim() || sending}
            className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white p-4 rounded-2xl disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-violet-500/20 hover:shadow-violet-500/40 active:scale-95"
          >
            <Send className="w-5 h-5" />
          </button>
        </form>
        <div className="text-center mt-3">
           <span className="text-xs text-slate-400 flex items-center justify-center gap-1.5">
             <Sparkles className="w-3 h-3 text-violet-400" /> {t('chat.aiDisclaimer')}
           </span>
        </div>
      </div>
    </div>
  );
};

// Also import MessageSquare for the empty state
import { MessageSquare } from 'lucide-react';
