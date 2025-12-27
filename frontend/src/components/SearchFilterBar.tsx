import React, { useState } from 'react';
import { Search, Filter, X, ChevronDown, Check } from 'lucide-react';
import { cn } from '../lib/utils';

export interface FilterGroup {
  id: string;
  label: string;
  options: { label: string; value: string }[]; // Updated to support label/value pairs
  selected: string[]; // Currently selected values
  onToggle: (value: string) => void;
}

interface SearchFilterBarProps {
  searchTerm: string;
  onSearchChange: (term: string) => void;
  filterGroups: FilterGroup[];
  placeholder?: string;
}

export const SearchFilterBar: React.FC<SearchFilterBarProps> = ({
  searchTerm,
  onSearchChange,
  filterGroups,
  placeholder = "Search..."
}) => {
  const [isFilterOpen, setIsFilterOpen] = useState(false);

  // Calculate total active filters
  const activeFilterCount = filterGroups.reduce((acc, group) => acc + group.selected.length, 0);

  return (
    <div className="flex gap-4">
      {/* Search Input */}
      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
        <input
          type="text"
          value={searchTerm}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder={placeholder}
          className="w-full pl-10 pr-4 py-2.5 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-400 shadow-sm"
        />
        {searchTerm && (
          <button 
            onClick={() => onSearchChange('')}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Extensible Filter Dropdown */}
      <div className="relative">
        <button
          onClick={() => setIsFilterOpen(!isFilterOpen)}
          className={cn(
            "h-full flex items-center gap-2 px-4 py-2.5 bg-white dark:bg-slate-800 border rounded-xl transition-all shadow-sm",
            isFilterOpen || activeFilterCount > 0
              ? "border-blue-500 text-blue-600 dark:text-blue-400 ring-2 ring-blue-500/20"
              : "border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:border-slate-300"
          )}
        >
          <Filter className="w-4 h-4" />
          <span className="text-sm font-medium">Filter</span>
          {activeFilterCount > 0 && (
            <span className="bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 text-xs px-2 py-0.5 rounded-full">
              {activeFilterCount}
            </span>
          )}
          <ChevronDown className="w-4 h-4" />
        </button>

        {isFilterOpen && (
          <>
            <div 
              className="fixed inset-0 z-20" 
              onClick={() => setIsFilterOpen(false)}
            />
            <div className="absolute right-0 top-full mt-2 w-64 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl z-30 p-2 max-h-[500px] overflow-y-auto">
              {filterGroups.map((group, groupIdx) => (
                <div key={group.id} className={cn(groupIdx > 0 && "mt-4 pt-4 border-t border-slate-100 dark:border-slate-800")}>
                  <div className="flex items-center justify-between px-2 py-1 mb-1">
                    <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">{group.label}</span>
                    {group.selected.length > 0 && (
                      <span className="text-xs text-blue-500 font-medium">{group.selected.length} selected</span>
                    )}
                  </div>
                  <div className="space-y-0.5">
                    {group.options.map(option => {
                      const isSelected = group.selected.includes(option.value);
                      return (
                        <button
                          key={option.value}
                          onClick={() => group.onToggle(option.value)}
                          className={cn(
                            "w-full flex items-center justify-between px-3 py-2 text-sm text-left rounded-lg transition-colors",
                            isSelected 
                              ? "bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300"
                              : "text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800"
                          )}
                        >
                          <span className="truncate">{option.label}</span>
                          {isSelected && <Check className="w-4 h-4 text-blue-500 flex-shrink-0" />}
                        </button>
                      );
                    })}
                  </div>
                </div>
              ))}
              
              {activeFilterCount > 0 && (
                <div className="border-t border-slate-100 dark:border-slate-800 mt-3 pt-2">
                  <button
                    onClick={() => {
                      filterGroups.forEach(g => {
                        // This relies on parent clearing logic, but typically we trigger individual clears.
                        // Ideally, we'd have a 'onClearAll' prop, but for now we rely on user clicking.
                        // Actually, let's just emit 'clear' intention if we redesign props, 
                        // but sticking to current contract: manual deselect or parent controlled.
                        // For this generic component, we might just hide this or loop toggle.
                        // Let's hide 'Clear All' here for simplicity or implement it if props allow.
                        // Since we don't have a global clear handler, we will just iterate toggle active ones.
                        g.selected.forEach(s => g.onToggle(s));
                      });
                      setIsFilterOpen(false);
                    }}
                    className="w-full text-center text-xs text-red-500 hover:text-red-600 py-2 hover:bg-red-50 dark:hover:bg-red-900/10 rounded-lg transition-colors"
                  >
                    Clear all filters
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
};