package agent

import "github.com/bitcoff/wtf/internal/provider"

// Имена инструментов — стабильные строки, в коде используем эти константы.
const (
	ToolRunCommand  = "run_command"
	ToolShowCommand = "show_command"
	ToolFinish      = "finish"

	// Используется только в режиме консолидации памяти, не в обычном цикле.
	ToolSaveConsolidated = "save_consolidated_memory"
)

// Tools — определения инструментов для главного цикла агента.
// Возвращаем функцией а не глобальной переменной чтобы избежать гонок при
// случайной мутации схемы.
func Tools() []provider.Tool {
	return []provider.Tool{
		{
			Name: ToolRunCommand,
			Description: "Запустить безопасную read-only команду на машине пользователя и получить её stdout/stderr. " +
				"Используй для диагностики: статусы сервисов, логи, ls, cat, ps, df, ip, и т.п. " +
				"Команда должна быть НЕ модифицирующей: никаких sudo, rm, restart, install, mkfs, dd. " +
				"Если команда destructive — система откажет, используй show_command вместо этого.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Полная команда с аргументами, например 'systemctl status nginx -l' или 'journalctl -u nginx -n 50'",
					},
					"reason": map[string]any{
						"type":        "string",
						"description": "Одна короткая фраза зачем эта команда нужна (показывается юзеру).",
					},
				},
				"required": []string{"command", "reason"},
			},
		},
		{
			Name: ToolShowCommand,
			Description: "Показать пользователю destructive-команду которую он должен выполнить сам. " +
				"Используй для sudo-команд, install/remove, restart сервисов, изменения конфигов. " +
				"После показа жди что юзер либо выполнит и пришлёт вывод, либо завершит сессию.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Точная команда без плейсхолдеров.",
					},
					"reason": map[string]any{
						"type":        "string",
						"description": "Что эта команда сделает и почему она нужна (1-2 строки).",
					},
				},
				"required": []string{"command", "reason"},
			},
		},
		{
			Name: ToolFinish,
			Description: "Завершить сессию. Вызывай когда диагностика завершена и юзеру дан ответ, " +
				"или когда не хватает данных и нужно честно об этом сказать.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"summary": map[string]any{
						"type":        "string",
						"description": "Финальный ответ юзеру: что было не так и что делать. Markdown допустим.",
					},
					"notes": map[string]any{
						"type": "array",
						"description": "Факты которые стоит запомнить надолго (опционально). " +
							"Каждая запись — короткая (до 200 символов).",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"type": map[string]any{
									"type": "string",
									"enum": []string{
										"machine_fact",
										"service_state",
										"user_preference",
										"resolved_issue",
									},
								},
								"key": map[string]any{
									"type":        "string",
									"description": "Уникальный идентификатор для дедупа, например 'nginx_version' или 'ssh_port'.",
								},
								"content": map[string]any{
									"type":        "string",
									"description": "Сам факт или состояние, до 200 символов.",
								},
								"ttl_days": map[string]any{
									"type":        "integer",
									"description": "Через сколько дней забыть. 0 = никогда. Для resolved_issue обычно 30, для machine_fact — 0.",
								},
							},
							"required": []string{"type", "key", "content"},
						},
					},
				},
				"required": []string{"summary"},
			},
		},
	}
}

// ConsolidationTools — отдельный набор для режима сжатия памяти.
// AI получает список старых entries и возвращает новый сжатый список.
func ConsolidationTools() []provider.Tool {
	return []provider.Tool{
		{
			Name:        ToolSaveConsolidated,
			Description: "Сохранить сжатый список записей памяти. Старые записи будут полностью заменены.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"entries": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"type": map[string]any{
									"type": "string",
									"enum": []string{
										"machine_fact",
										"service_state",
										"user_preference",
										"resolved_issue",
									},
								},
								"key":      map[string]any{"type": "string"},
								"content":  map[string]any{"type": "string"},
								"ttl_days": map[string]any{"type": "integer"},
							},
							"required": []string{"type", "key", "content"},
						},
					},
				},
				"required": []string{"entries"},
			},
		},
	}
}
