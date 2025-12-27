import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useLanguage } from '../context/LanguageContext';
import { mockApi } from '../services/mockApi';
import { CreditCard, History, User, Brain, Shield, Trash2, Key, AlertTriangle, Loader2 } from 'lucide-react';
import { cn } from '../lib/utils';

export const ProfilePage: React.FC = () => {
  const { user, logout } = useAuth();
  const { t } = useLanguage();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<'overview' | 'memories' | 'settings'>('overview');
  
  // Memories State
  const [memories, setMemories] = useState<string[]>([]);
  const [loadingMemories, setLoadingMemories] = useState(false);

  // Settings State
  const [passwords, setPasswords] = useState({ old: '', new: '' });
  const [loadingSettings, setLoadingSettings] = useState(false);
  const [verificationStep, setVerificationStep] = useState<'idle' | 'verifying' | 'verified'>('idle');
  const [verificationCode, setVerificationCode] = useState('');
  const [pendingAction, setPendingAction] = useState<'password' | 'delete' | null>(null);

  const startVerification = (action: 'password' | 'delete') => {
    setPendingAction(action);
    setVerificationStep('verifying');
  };

  const handleSendCode = async () => {
    if (!user) return;
    const contact = user.email || user.phone || user.name;
    await mockApi.security.sendVerificationCode(contact);
    alert(t('auth.codeSent'));
  };

  const handleVerifyCode = async () => {
    if (!user) return;
    const contact = user.email || user.phone || user.name;
    try {
      setLoadingSettings(true);
      await mockApi.security.verifyCode(contact, verificationCode);
      setVerificationStep('verified');
    } catch (error) {
      alert(t('auth.codeInvalid'));
    } finally {
      setLoadingSettings(false);
    }
  };

  const handleUpdatePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (verificationStep !== 'verified') {
      startVerification('password');
      return;
    }
    setLoadingSettings(true);
    try {
      await mockApi.user.updatePassword(passwords.old, passwords.new);
      alert('Password updated successfully');
      setPasswords({ old: '', new: '' });
      setVerificationStep('idle');
    } catch (error) {
      alert('Failed to update password');
    } finally {
      setLoadingSettings(false);
    }
  };

  const handleDeleteAccount = async () => {
    if (verificationStep !== 'verified') {
      startVerification('delete');
      return;
    }
    if (window.confirm(t('profile.settings.confirmDelete'))) {
      try {
        await mockApi.user.deleteAccount();
        await logout();
        navigate('/login');
      } catch (error) {
        console.error(error);
      }
    }
  };

  const tabs = [
    { id: 'overview', label: t('profile.tab.overview'), icon: User },
    { id: 'memories', label: t('profile.tab.memories'), icon: Brain },
    { id: 'settings', label: t('profile.tab.settings'), icon: Shield },
  ];

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <h1 className="text-3xl font-bold text-slate-900 dark:text-white">{t('profile.title')}</h1>

      <div className="grid md:grid-cols-12 gap-6">
        {/* Left Column: User Card & Navigation */}
        <div className="md:col-span-4 space-y-6">
          <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-800 flex flex-col items-center text-center">
            <img src={user?.avatar} className="w-24 h-24 rounded-full mb-4 object-cover" alt="Profile" />
            <h2 className="text-xl font-bold text-slate-900 dark:text-white">{user?.name}</h2>
            <p className="text-slate-500 dark:text-slate-400 text-sm">{user?.email}</p>
            <div className="mt-4 px-3 py-1 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 rounded-full text-xs font-semibold">
              {t('profile.premium')}
            </div>
          </div>

          <nav className="space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={cn(
                  "w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium text-sm",
                  activeTab === tab.id
                    ? "bg-white dark:bg-slate-800 text-blue-600 dark:text-blue-400 shadow-sm"
                    : "text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800/50 hover:text-slate-900 dark:hover:text-slate-200"
                )}
              >
                <tab.icon className="w-5 h-5" />
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Right Column: Content Area */}
        <div className="md:col-span-8">
          {activeTab === 'overview' && (
            <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
              <div className="bg-gradient-to-r from-slate-900 to-slate-800 dark:from-blue-900 dark:to-slate-900 rounded-2xl p-6 text-white shadow-lg">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-semibold flex items-center gap-2">
                    <CreditCard className="w-5 h-5 text-cyan-400" /> 
                    {t('profile.balance')}
                  </h3>
                  <span className="text-xs bg-white/10 px-2 py-1 rounded">{t('profile.autoRefill')}</span>
                </div>
                <div className="text-4xl font-bold mb-2">{user?.credits} <span className="text-lg font-normal text-slate-400">credits</span></div>
                <p className="text-slate-400 text-sm">~{user?.credits && user.credits * 10} {t('profile.remaining')}</p>
                
                <div className="mt-6 flex gap-3">
                   <button className="bg-cyan-500 hover:bg-cyan-400 text-white px-4 py-2 rounded-lg text-sm font-medium transition">
                     {t('profile.buy')}
                   </button>
                   <button className="bg-white/10 hover:bg-white/20 text-white px-4 py-2 rounded-lg text-sm font-medium transition">
                     {t('profile.history')}
                   </button>
                </div>
              </div>

              <div className="bg-white dark:bg-slate-900 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-800 overflow-hidden">
                <div className="p-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between">
                  <h3 className="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <History className="w-4 h-4" /> {t('profile.recentActivity')}
                  </h3>
                </div>
                <div className="p-4">
                  <div className="space-y-4">
                    {[1, 2, 3].map((i) => (
                      <div key={i} className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-3">
                          <div className="w-2 h-2 rounded-full bg-slate-300 dark:bg-slate-600"></div>
                          <span className="text-slate-600 dark:text-slate-300">{t('profile.sessionWith')} #{i}</span>
                        </div>
                        <span className="text-slate-400 dark:text-slate-500">2 {t('profile.ago')}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'memories' && (
            <div className="bg-white dark:bg-slate-900 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-800 p-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
              <div className="mb-6">
                <h3 className="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Brain className="w-5 h-5 text-purple-500" />
                  {t('profile.memories.title')}
                </h3>
                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">{t('profile.memories.desc')}</p>
              </div>

              {loadingMemories ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="w-8 h-8 animate-spin text-purple-500" />
                </div>
              ) : memories.length === 0 ? (
                <div className="text-center py-12 bg-slate-50 dark:bg-slate-800/50 rounded-xl border border-dashed border-slate-200 dark:border-slate-700">
                  <p className="text-slate-500 dark:text-slate-400">{t('profile.memories.empty')}</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {memories.map((memory, idx) => (
                    <div key={idx} className="flex items-start justify-between p-4 bg-slate-50 dark:bg-slate-800/50 rounded-xl group border border-transparent hover:border-purple-200 dark:hover:border-purple-800 transition-all">
                      <p className="text-slate-700 dark:text-slate-300 text-sm leading-relaxed">{memory}</p>
                      <button 
                        onClick={() => handleDeleteMemory(idx)}
                        className="text-slate-400 hover:text-red-500 p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                        title="Delete memory"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'settings' && (
            <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
              {verificationStep === 'verifying' ? (
                <div className="bg-white dark:bg-slate-900 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-800 p-6">
                  <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2">
                    <Shield className="w-5 h-5 text-blue-500" />
                    {t('auth.verifyTitle')}
                  </h3>
                  <p className="text-slate-500 dark:text-slate-400 text-sm mb-6">{t('auth.verifyDesc')}</p>
                  
                  <div className="space-y-4 max-w-md">
                    <div className="flex gap-2">
                      <input 
                        type="text"
                        value={verificationCode}
                        onChange={(e) => setVerificationCode(e.target.value)}
                        placeholder={t('auth.code')}
                        className="flex-1 px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 outline-none"
                      />
                      <button
                        type="button"
                        onClick={handleSendCode}
                        className="bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-4 py-2 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-slate-700 transition"
                      >
                        {t('auth.sendCode')}
                      </button>
                    </div>
                    <div className="flex gap-2">
                      <button 
                        type="button" 
                        onClick={handleVerifyCode}
                        disabled={loadingSettings || !verificationCode}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition disabled:opacity-50 flex-1"
                      >
                        {loadingSettings ? <Loader2 className="w-4 h-4 animate-spin mx-auto" /> : t('auth.verify')}
                      </button>
                      <button 
                        type="button" 
                        onClick={() => {
                          setVerificationStep('idle');
                          setVerificationCode('');
                        }}
                        className="bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-4 py-2 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-slate-700 transition"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </div>
              ) : (
                <>
                  <div className="bg-white dark:bg-slate-900 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-800 p-6">
                    <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-6 flex items-center gap-2">
                      <Key className="w-5 h-5 text-blue-500" />
                      {t('profile.settings.password')}
                    </h3>
                    <form onSubmit={handleUpdatePassword} className="space-y-4 max-w-md">
                      <div>
                        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{t('profile.settings.oldPass')}</label>
                        <input 
                          type="password"
                          value={passwords.old}
                          onChange={e => setPasswords({...passwords, old: e.target.value})}
                          className="w-full px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 outline-none"
                          disabled={verificationStep !== 'verified' && pendingAction === 'password'}
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">{t('profile.settings.newPass')}</label>
                        <input 
                          type="password"
                          value={passwords.new}
                          onChange={e => setPasswords({...passwords, new: e.target.value})}
                          className="w-full px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 outline-none"
                          disabled={verificationStep !== 'verified' && pendingAction === 'password'}
                        />
                      </div>
                      <button 
                        type="submit" 
                        disabled={loadingSettings}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition disabled:opacity-50"
                      >
                        {verificationStep === 'verified' && pendingAction === 'password' ? t('profile.settings.update') : t('auth.verifyTitle')}
                      </button>
                    </form>
                  </div>

                  <div className="bg-red-50 dark:bg-red-900/10 rounded-2xl shadow-sm border border-red-100 dark:border-red-900/30 p-6">
                    <h3 className="text-lg font-bold text-red-600 dark:text-red-400 mb-2 flex items-center gap-2">
                      <AlertTriangle className="w-5 h-5" />
                      {t('profile.settings.danger')}
                    </h3>
                    <p className="text-red-600/70 dark:text-red-400/70 text-sm mb-6">
                      {t('profile.settings.deleteDesc')}
                    </p>
                    <button 
                      onClick={handleDeleteAccount}
                      className="bg-red-600 hover:bg-red-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition"
                    >
                      {verificationStep === 'verified' && pendingAction === 'delete' ? t('profile.settings.deleteAccount') : t('auth.verifyTitle')}
                    </button>
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};