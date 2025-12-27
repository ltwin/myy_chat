import React, { useState, useRef, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { mockApi } from '../services/mockApi';
import { useLanguage } from '../context/LanguageContext';
import { Loader2, Save, Upload, X, ChevronDown, Check } from 'lucide-react';
import { cn } from '../lib/utils';

export const CreatePage: React.FC = () => {
  const navigate = useNavigate();
  const { charId } = useParams<{ charId: string }>();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { t } = useLanguage();
  const [loading, setLoading] = useState(false);
  const [tagDropdownOpen, setTagDropdownOpen] = useState(false);
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    personality: '',
    tags: [] as string[],
    avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${Date.now()}`
  });

  useEffect(() => {
    const init = async () => {
      try {
        const tags = await mockApi.system.getTags();
        setAvailableTags(tags);
      } catch (err) {
        console.error(err);
      }
    };
    init();
  }, []);

  useEffect(() => {
    if (charId) {
      const loadChar = async () => {
        setLoading(true);
        try {
          const char = await mockApi.character.get(charId);
          if (char) {
            setFormData({
              name: char.name,
              description: char.description,
              personality: char.personality,
              tags: char.tags,
              avatar: char.avatar
            });
          }
        } finally {
          setLoading(false);
        }
      };
      loadChar();
    }
  }, [charId]);

  const toggleTag = (tag: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.includes(tag) 
        ? prev.tags.filter(t => t !== tag)
        : [...prev.tags, tag]
    }));
  };

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        setFormData(prev => ({ ...prev, avatar: reader.result as string }));
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const data = {
        ...formData,
        avatar: formData.avatar
      };

      if (charId) {
        await mockApi.character.update(charId, data);
      } else {
        await mockApi.character.create(data);
      }
      navigate('/');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
          {charId ? t('create.editTitle') : t('create.title')}
        </h1>
        <p className="text-slate-500 dark:text-slate-400 mt-2">
          {charId ? t('create.editSubtitle') : t('create.subtitle')}
        </p>
      </div>

      <div className="bg-white dark:bg-slate-900 rounded-2xl p-8 shadow-sm border border-slate-100 dark:border-slate-800 transition-colors">
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="flex flex-col items-center mb-6 gap-3">
            <div 
              className="relative group cursor-pointer"
              onClick={() => fileInputRef.current?.click()}
            >
              <img 
                src={formData.avatar} 
                alt="Avatar Preview" 
                className="w-24 h-24 rounded-full bg-slate-100 dark:bg-slate-800 object-cover border-4 border-white dark:border-slate-700 shadow-md group-hover:opacity-75 transition-opacity"
              />
              <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                <Upload className="w-8 h-8 text-white drop-shadow-lg" />
              </div>
            </div>
            <span className="text-xs text-slate-500 dark:text-slate-400">{t('create.uploadAvatar')}</span>
            <input 
              type="file" 
              ref={fileInputRef} 
              className="hidden" 
              accept="image/*"
              onChange={handleImageUpload}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{t('create.name')}</label>
            <input
              required
              className="w-full px-4 py-2 border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition-colors"
              value={formData.name}
              onChange={e => setFormData({...formData, name: e.target.value})}
              placeholder={t('create.placeholder.name')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{t('create.description')}</label>
            <input
              required
              className="w-full px-4 py-2 border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition-colors"
              value={formData.description}
              onChange={e => setFormData({...formData, description: e.target.value})}
              placeholder={t('create.placeholder.desc')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{t('create.personality')}</label>
            <textarea
              required
              rows={3}
              className="w-full px-4 py-2 border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg focus:ring-2 focus:ring-blue-500 outline-none resize-none transition-colors"
              value={formData.personality}
              onChange={e => setFormData({...formData, personality: e.target.value})}
              placeholder={t('create.placeholder.pers')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{t('create.tags')}</label>
            <div className="relative">
              <button
                type="button"
                onClick={() => setTagDropdownOpen(!tagDropdownOpen)}
                className="w-full flex items-center justify-between px-4 py-2.5 border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-left min-h-[44px]"
              >
                <div className="flex flex-wrap gap-2">
                  {formData.tags.length === 0 ? (
                    <span className="text-slate-400">{t('create.placeholder.tags')}</span>
                  ) : (
                    formData.tags.map(tag => (
                      <span key={tag} className="inline-flex items-center gap-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 px-2 py-0.5 rounded text-xs font-medium">
                        {t(`tag.${tag.toLowerCase()}` as any)}
                        <X 
                          className="w-3 h-3 hover:text-blue-900 cursor-pointer" 
                          onClick={(e) => { e.stopPropagation(); toggleTag(tag); }} 
                        />
                      </span>
                    ))
                  )}
                </div>
                <ChevronDown className="w-4 h-4 text-slate-400 ml-2 flex-shrink-0" />
              </button>

              {tagDropdownOpen && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setTagDropdownOpen(false)} />
                  <div className="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg shadow-xl z-20 max-h-60 overflow-y-auto p-2 grid grid-cols-2 gap-1">
                    {availableTags.map(tag => {
                      const isSelected = formData.tags.includes(tag);
                      return (
                        <button
                          key={tag}
                          type="button"
                          onClick={() => toggleTag(tag)}
                          className={cn(
                            "flex items-center justify-between px-3 py-2 rounded-md text-sm text-left transition-colors",
                            isSelected 
                              ? "bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400"
                              : "text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800"
                          )}
                        >
                          <span>{t(`tag.${tag.toLowerCase()}` as any)}</span>
                          {isSelected && <Check className="w-4 h-4" />}
                        </button>
                      );
                    })}
                  </div>
                </>
              )}
            </div>
          </div>

          <div className="pt-4">
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white py-3.5 rounded-xl font-semibold transition-all shadow-lg shadow-violet-500/25 hover:shadow-violet-500/40 hover:-translate-y-0.5 flex items-center justify-center gap-2"
            >
              {loading ? <Loader2 className="animate-spin" /> : <Save className="w-5 h-5" />}
              {charId ? t('create.save') : t('create.submit')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
