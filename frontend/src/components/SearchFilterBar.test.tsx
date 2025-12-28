/**
 * SearchFilterBar 组件单元测试
 * 测试搜索输入和过滤器下拉功能
 */
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { SearchFilterBar, type FilterGroup } from './SearchFilterBar'

// 测试用的模拟过滤器组数据
const createMockFilterGroups = (overrides?: Partial<FilterGroup>[]): FilterGroup[] => [
  {
    id: 'category',
    label: 'Category',
    options: [
      { label: 'Option A', value: 'a' },
      { label: 'Option B', value: 'b' },
      { label: 'Option C', value: 'c' },
    ],
    selected: [],
    onToggle: vi.fn(),
    ...overrides?.[0],
  },
  {
    id: 'status',
    label: 'Status',
    options: [
      { label: 'Active', value: 'active' },
      { label: 'Inactive', value: 'inactive' },
    ],
    selected: [],
    onToggle: vi.fn(),
    ...overrides?.[1],
  },
]

describe('SearchFilterBar', () => {
  describe('搜索功能', () => {
    it('渲染搜索输入框并显示占位符', () => {
      const onSearchChange = vi.fn()
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={onSearchChange}
          filterGroups={createMockFilterGroups()}
          placeholder="Search characters..."
        />
      )

      const searchInput = screen.getByPlaceholderText('Search characters...')
      expect(searchInput).toBeInTheDocument()
    })

    it('显示当前搜索词', () => {
      const onSearchChange = vi.fn()
      render(
        <SearchFilterBar
          searchTerm="test query"
          onSearchChange={onSearchChange}
          filterGroups={createMockFilterGroups()}
        />
      )

      const searchInput = screen.getByDisplayValue('test query')
      expect(searchInput).toBeInTheDocument()
    })

    it('输入时调用 onSearchChange', () => {
      const onSearchChange = vi.fn()
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={onSearchChange}
          filterGroups={createMockFilterGroups()}
        />
      )

      const searchInput = screen.getByRole('textbox')
      fireEvent.change(searchInput, { target: { value: 'new search' } })

      expect(onSearchChange).toHaveBeenCalledWith('new search')
    })

    it('有搜索词时显示清除按钮', () => {
      const onSearchChange = vi.fn()
      render(
        <SearchFilterBar
          searchTerm="some text"
          onSearchChange={onSearchChange}
          filterGroups={createMockFilterGroups()}
        />
      )

      // 清除按钮应该存在（X 图标按钮）
      const clearButtons = screen.getAllByRole('button')
      // 应该有 Filter 按钮和清除按钮
      expect(clearButtons.length).toBeGreaterThanOrEqual(2)
    })

    it('点击清除按钮时清空搜索词', () => {
      const onSearchChange = vi.fn()
      render(
        <SearchFilterBar
          searchTerm="some text"
          onSearchChange={onSearchChange}
          filterGroups={createMockFilterGroups()}
        />
      )

      // 找到清除按钮（搜索框内的 X 按钮）
      const buttons = screen.getAllByRole('button')
      // 第一个应该是清除按钮（在搜索框内）
      const clearButton = buttons[0]
      fireEvent.click(clearButton)

      expect(onSearchChange).toHaveBeenCalledWith('')
    })
  })

  describe('过滤器功能', () => {
    it('渲染 Filter 按钮', () => {
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={createMockFilterGroups()}
        />
      )

      expect(screen.getByText('Filter')).toBeInTheDocument()
    })

    it('点击 Filter 按钮打开下拉菜单', () => {
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={createMockFilterGroups()}
        />
      )

      const filterButton = screen.getByText('Filter').closest('button')!
      fireEvent.click(filterButton)

      // 应该显示过滤器组标签
      expect(screen.getByText('Category')).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
    })

    it('显示过滤器选项', () => {
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={createMockFilterGroups()}
        />
      )

      // 打开过滤器下拉
      const filterButton = screen.getByText('Filter').closest('button')!
      fireEvent.click(filterButton)

      // 验证选项存在
      expect(screen.getByText('Option A')).toBeInTheDocument()
      expect(screen.getByText('Option B')).toBeInTheDocument()
      expect(screen.getByText('Active')).toBeInTheDocument()
      expect(screen.getByText('Inactive')).toBeInTheDocument()
    })

    it('点击选项时调用 onToggle', () => {
      const mockGroups = createMockFilterGroups()
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={mockGroups}
        />
      )

      // 打开过滤器下拉
      const filterButton = screen.getByText('Filter').closest('button')!
      fireEvent.click(filterButton)

      // 点击选项
      const optionButton = screen.getByText('Option A')
      fireEvent.click(optionButton)

      expect(mockGroups[0].onToggle).toHaveBeenCalledWith('a')
    })

    it('显示已选中过滤器的计数徽章', () => {
      const mockGroups = createMockFilterGroups([
        { selected: ['a', 'b'] }, // 2 selected in category
        { selected: ['active'] }, // 1 selected in status
      ])

      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={mockGroups}
        />
      )

      // 应该显示总计数 (3)
      expect(screen.getByText('3')).toBeInTheDocument()
    })

    it('有选中过滤器时显示 Clear all filters 按钮', () => {
      const mockGroups = createMockFilterGroups([
        { selected: ['a'] },
        { selected: [] },
      ])

      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={mockGroups}
        />
      )

      // 打开过滤器下拉
      const filterButton = screen.getByText('Filter').closest('button')!
      fireEvent.click(filterButton)

      expect(screen.getByText('Clear all filters')).toBeInTheDocument()
    })

    it('点击 Clear all filters 清除所有选中', () => {
      const mockGroups = createMockFilterGroups([
        { selected: ['a', 'b'] },
        { selected: ['active'] },
      ])

      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={mockGroups}
        />
      )

      // 打开过滤器下拉
      const filterButton = screen.getByText('Filter').closest('button')!
      fireEvent.click(filterButton)

      // 点击清除所有
      const clearAllButton = screen.getByText('Clear all filters')
      fireEvent.click(clearAllButton)

      // 应该对每个已选中的值调用 onToggle
      expect(mockGroups[0].onToggle).toHaveBeenCalledWith('a')
      expect(mockGroups[0].onToggle).toHaveBeenCalledWith('b')
      expect(mockGroups[1].onToggle).toHaveBeenCalledWith('active')
    })
  })

  describe('默认值', () => {
    it('使用默认占位符', () => {
      render(
        <SearchFilterBar
          searchTerm=""
          onSearchChange={vi.fn()}
          filterGroups={createMockFilterGroups()}
        />
      )

      expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument()
    })
  })
})
