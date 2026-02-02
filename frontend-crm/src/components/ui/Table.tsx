// ============================================
// Table Component
// Production-Ready Data Table
// ============================================

import { type ReactNode } from 'react';

interface Column<T> {
    key: keyof T | string;
    header: string;
    width?: string;
    align?: 'left' | 'center' | 'right';
    render?: (item: T, index: number) => ReactNode;
}

interface TableProps<T> {
    columns: Column<T>[];
    data: T[];
    keyExtractor: (item: T) => string;
    loading?: boolean;
    emptyMessage?: string;
    onRowClick?: (item: T) => void;
    selectedRows?: string[];
    onSelectionChange?: (selectedIds: string[]) => void;
    showCheckboxes?: boolean;
}

export function Table<T>({
    columns,
    data,
    keyExtractor,
    loading = false,
    emptyMessage = 'No data available',
    onRowClick,
    selectedRows = [],
    onSelectionChange,
    showCheckboxes = false,
}: TableProps<T>) {
    const allSelected = data.length > 0 && selectedRows.length === data.length;
    const someSelected = selectedRows.length > 0 && selectedRows.length < data.length;

    const handleSelectAll = () => {
        if (onSelectionChange) {
            if (allSelected) {
                onSelectionChange([]);
            } else {
                onSelectionChange(data.map(keyExtractor));
            }
        }
    };

    const handleSelectRow = (id: string) => {
        if (onSelectionChange) {
            if (selectedRows.includes(id)) {
                onSelectionChange(selectedRows.filter((rowId) => rowId !== id));
            } else {
                onSelectionChange([...selectedRows, id]);
            }
        }
    };

    const getCellValue = (item: T, column: Column<T>): ReactNode => {
        if (column.render) {
            return column.render(item, data.indexOf(item));
        }
        const value = (item as Record<string, unknown>)[column.key as string];
        return value as ReactNode;
    };

    if (loading) {
        return (
            <div className="table-wrapper">
                <table className="table">
                    <thead>
                        <tr>
                            {showCheckboxes && (
                                <th style={{ width: '50px' }}>
                                    <div className="skeleton" style={{ width: '20px', height: '20px' }} />
                                </th>
                            )}
                            {columns.map((column) => (
                                <th key={String(column.key)} style={{ width: column.width }}>
                                    {column.header}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {[...Array(5)].map((_, index) => (
                            <tr key={index}>
                                {showCheckboxes && (
                                    <td>
                                        <div className="skeleton" style={{ width: '20px', height: '20px' }} />
                                    </td>
                                )}
                                {columns.map((column) => (
                                    <td key={String(column.key)}>
                                        <div className="skeleton" style={{ width: '100%', height: '20px' }} />
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        );
    }

    if (data.length === 0) {
        return (
            <div className="table-wrapper">
                <div className="empty-state">
                    <div className="empty-state-icon">📋</div>
                    <h3 className="empty-state-title">{emptyMessage}</h3>
                </div>
            </div>
        );
    }

    return (
        <div className="table-wrapper">
            <table className="table">
                <thead>
                    <tr>
                        {showCheckboxes && (
                            <th style={{ width: '50px' }}>
                                <input
                                    type="checkbox"
                                    checked={allSelected}
                                    ref={(el) => {
                                        if (el) el.indeterminate = someSelected;
                                    }}
                                    onChange={handleSelectAll}
                                    className="form-checkbox"
                                />
                            </th>
                        )}
                        {columns.map((column) => (
                            <th
                                key={String(column.key)}
                                style={{ width: column.width, textAlign: column.align || 'left' }}
                            >
                                {column.header}
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {data.map((item) => {
                        const id = keyExtractor(item);
                        const isSelected = selectedRows.includes(id);

                        return (
                            <tr
                                key={id}
                                className={`${isSelected ? 'selected' : ''} ${onRowClick ? 'cursor-pointer' : ''}`}
                                onClick={() => onRowClick?.(item)}
                            >
                                {showCheckboxes && (
                                    <td onClick={(e) => e.stopPropagation()}>
                                        <input
                                            type="checkbox"
                                            checked={isSelected}
                                            onChange={() => handleSelectRow(id)}
                                            className="form-checkbox"
                                        />
                                    </td>
                                )}
                                {columns.map((column) => (
                                    <td
                                        key={String(column.key)}
                                        style={{ textAlign: column.align || 'left' }}
                                    >
                                        {getCellValue(item, column)}
                                    </td>
                                ))}
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}

export default Table;
