import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import type { Character } from '../types';
import { mockApi } from '../services/mockApi';
import { useChat } from '../context/ChatContext';
import { useLanguage } from '../context/LanguageContext';
import { SearchFilterBar } from '../components/SearchFilterBar';
import { MessageSquare, Copy, Loader2, Sparkles, Globe } from 'lucide-react';

export const MarketplacePage: React.FC = () => {
  const [characters, setCharacters] = useState<Character[]>([]);
  const [loading, setLoading] = useState(true);
  const [forking, setForking] = useState<string | null>(null);
  const { createSession } = useChat();
  const { t } = useLanguage();
  const navigate = useNavigate();

  // Search & Filter State
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [availableTags, setAvailableTags] = useState<string[]>([]);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [chars, tags] = await Promise.all([
          mockApi.character.listPublic(),
          mockApi.system.getTags()
        ]);
        setCharacters(chars);
        setAvailableTags(tags);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  const handleStartChat = async (charId: string) => {
    try {
      const sessionId = await createSession(charId);
      navigate(`/chat/${sessionId}`);
    } catch (error) {
      console.error("Failed to start chat", error);
    }
  };

        // Filter logic
        const filteredCharacters = characters.filter(char => {
          const matchesSearch = char.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                                char.description.toLowerCase().includes(searchTerm.toLowerCase());
          
          // Case-insensitive tag matching
          const matchesTags = selectedTags.length === 0 || selectedTags.every(selectedTag => 
            char.tags.some(charTag => charTag.toLowerCase() === selectedTag.toLowerCase())
          );
          
          return matchesSearch && matchesTags;
        });  const handleTagToggle = (tag: string) => {
    setSelectedTags(prev => 
      prev.includes(tag) ? prev.filter(t => t !== tag) : [...prev, tag]
    );
  };

  const filterGroups = [
    {
      id: 'tags',
      label: 'Tags',
      options: availableTags.map(tag => ({
        value: tag,
        label: t(`tag.${tag.toLowerCase()}` as any)
      })),
      selected: selectedTags,
      onToggle: handleTagToggle
    }
  ];

  const handleFork = async (e: React.MouseEvent, charId: string) => {
    e.stopPropagation();
    setForking(charId);
    try {
      await mockApi.character.fork(charId);
      navigate('/'); // Redirect to "My Companions"
    } catch (error) {
      console.error("Failed to fork character", error);
    } finally {
      setForking(null);
    }
  };

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-violet-500" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-extrabold bg-gradient-to-r from-cyan-600 to-blue-600 dark:from-cyan-400 dark:to-blue-400 bg-clip-text text-transparent flex items-center gap-3">
            <Globe className="w-8 h-8 text-blue-500" />
            {t('market.title')}
          </h1>
          <p className="text-slate-500 dark:text-slate-400 mt-2 text-lg">{t('market.subtitle')}</p>
        </div>
      </div>

      <SearchFilterBar 
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        filterGroups={filterGroups}
        placeholder={t('market.searchPlaceholder') || "Search market..."}
      />

      {filteredCharacters.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-slate-500 dark:text-slate-400">No characters found matching your criteria.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredCharacters.map((char) => (
            <div 
              key={char.id}
              className="group relative bg-white/70 dark:bg-slate-900/60 backdrop-blur-xl rounded-2xl border border-white/50 dark:border-slate-700/50 p-6 hover:border-cyan-500/50 dark:hover:border-cyan-500/50 hover:shadow-2xl hover:shadow-cyan-500/10 transition-all duration-300 cursor-pointer overflow-hidden"
              onClick={() => handleStartChat(char.id)}
            >
            <div className="absolute inset-0 bg-gradient-to-br from-cyan-500/5 to-blue-500/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>

            <button
              onClick={(e) => handleFork(e, char.id)}
              disabled={!!forking}
              className="absolute top-4 right-4 p-2 text-slate-400 hover:text-white hover:bg-cyan-500 rounded-lg opacity-0 group-hover:opacity-100 transition-all z-10 shadow-lg flex items-center gap-2 font-medium text-xs bg-white/80 dark:bg-slate-800/80 backdrop-blur"
              title="Add to My Companions"
            >
              {forking === char.id ? <Loader2 className="w-4 h-4 animate-spin" /> : <Copy className="w-4 h-4" />}
              <span className="hidden group-hover:inline">{t('market.fork')}</span>
            </button>

            <div className="flex items-start justify-between mb-5 relative z-0">
              <div className="relative">
                <img 
                  src={char.avatar} 
                  alt={char.name}
                  className="relative w-16 h-16 rounded-xl object-cover shadow-sm group-hover:scale-105 transition-transform duration-300 ring-2 ring-white/50 dark:ring-white/10" 
                />
              </div>
              <div className="flex gap-2">
                <span className="text-xs font-semibold px-2.5 py-1 bg-cyan-100/50 dark:bg-cyan-900/30 text-cyan-700 dark:text-cyan-300 rounded-full border border-cyan-200/50 dark:border-cyan-700/50 backdrop-blur-sm">
                  {char.author || 'Official'}
                </span>
                {char.tags.slice(0, 1).map(tag => (
                  <span key={tag} className="text-xs font-semibold px-2.5 py-1 bg-slate-100/80 dark:bg-slate-800/80 text-slate-600 dark:text-slate-300 rounded-full border border-slate-200/50 dark:border-slate-700/50 backdrop-blur-sm">
                    {t(`tag.${tag.toLowerCase()}` as any)}
                  </span>
                ))}
              </div>
            </div>
            
            <h3 className="relative z-0 text-xl font-bold text-slate-900 dark:text-white mb-2 group-hover:text-cyan-600 dark:group-hover:text-cyan-400 transition-colors">
              {char.name}
            </h3>
            <p className="relative z-0 text-sm text-slate-500 dark:text-slate-400 line-clamp-2 mb-6 h-10 leading-relaxed">
              {char.description}
            </p>

            <button className="relative z-0 w-full py-3 rounded-xl bg-slate-50/50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 font-semibold text-sm group-hover:bg-gradient-to-r group-hover:from-cyan-600 group-hover:to-blue-600 group-hover:text-white transition-all flex items-center justify-center gap-2 group-hover:shadow-lg group-hover:shadow-cyan-500/20">
              <MessageSquare className="w-4 h-4" />
              {t('dash.startChat')}
            </button>
          </div>
        ))}
        </div>
      )}
    </div>
  );
};
