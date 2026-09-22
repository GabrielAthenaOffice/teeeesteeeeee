import React, { useState } from 'react';
import './DataGrid.css';

export function DataGrid({ columns, data, selectable = true, onRowClick }) {
  const [selectedRows, setSelectedRows] = useState(new Set());

  const handleSelectAll = (e) => {
    if (e.target.checked) {
      setSelectedRows(new Set(data.map((_, i) => i)));
    } else {
      setSelectedRows(new Set());
    }
  };

  const handleSelectRow = (index) => {
    const newSelected = new Set(selectedRows);
    if (newSelected.has(index)) {
      newSelected.delete(index);
    } else {
      newSelected.add(index);
    }
    setSelectedRows(newSelected);
  };

  if (!data || data.length === 0) {
    return (
      <div className="datagrid-container">
        <div className="datagrid-empty">Nenhum registro encontrado.</div>
      </div>
    );
  }

  return (
    <div className="datagrid-container">
      <table className="datagrid">
        <thead>
          <tr>
            {selectable && (
              <th className="datagrid-checkbox">
                <input 
                  type="checkbox" 
                  onChange={handleSelectAll} 
                  checked={selectedRows.size === data.length && data.length > 0} 
                />
              </th>
            )}
            {columns.map((col, idx) => (
              <th key={idx} style={{ width: col.width || 'auto' }}>
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, rowIndex) => (
            <tr
              key={rowIndex}
              className={selectedRows.has(rowIndex) ? 'selected' : ''}
              style={onRowClick ? { cursor: 'pointer' } : undefined}
              onClick={() => {
                if (selectable) handleSelectRow(rowIndex);
                onRowClick?.(row);
              }}
            >
              {selectable && (
                <td className="datagrid-checkbox" onClick={e => e.stopPropagation()}>
                  <input 
                    type="checkbox" 
                    checked={selectedRows.has(rowIndex)}
                    onChange={() => handleSelectRow(rowIndex)}
                  />
                </td>
              )}
              {columns.map((col, colIndex) => (
                <td key={colIndex}>
                  {col.render ? col.render(row[col.field], row) : row[col.field]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
