// ============================================================
// I18N (INTERNATIONALIZATION)
// ============================================================
let currentLang = localStorage.getItem('librarr_lang') || 'en';

const I18N = {
  en: {
    // Navigation
    nav_search: 'Search',
    nav_library: 'Library',
    nav_browse: 'Browse',
    nav_downloads: 'Downloads',
    nav_wishlist: 'Wishlist',
    nav_settings: 'Settings',
    sign_out: 'Sign out',
    // Header
    header_subtitle: 'Self-hosted book manager',
    // Login modal
    login_subtitle: 'Sign in to continue',
    login_sign_in: 'Sign In',
    login_or: 'or',
    login_sso: 'Login with SSO',
    login_with: 'Login with {provider}',
    login_no_account: 'No account? <a href="#" data-action="showRegisterForm" class="text-indigo-400 hover:text-indigo-300">Register</a>',
    create_admin_account: 'Create your admin account',
    create_your_account: 'Create your account',
    create_account: 'Create Account',
    back_to_login: 'Back to login',
    // Labels
    label_username: 'Username',
    label_password: 'Password',
    // Placeholders
    ph_username: 'Username',
    ph_password: 'Password',
    ph_choose_username: 'Choose a username',
    ph_min_6_chars: 'Min 6 characters',
    // TOTP
    totp_enter_code: 'Enter the 6-digit code from your authenticator app, or a backup code.',
    totp_code_label: 'TOTP Code',
    totp_verify: 'Verify',
    s_totp_title: 'Two-Factor Authentication',
    s_totp_desc: 'Add an extra layer of security with a TOTP authenticator app.',
    enable_2fa: 'Enable 2FA',
    disable_2fa: 'Disable 2FA',
    totp_enabled_msg: 'Two-factor authentication is enabled.',
    totp_scan_qr: 'Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)',
    totp_manual_secret: 'Or enter this secret manually:',
    totp_backup_codes: 'Backup Codes (save these somewhere safe!):',
    verify_enable: 'Verify & Enable',
    confirm_disable: 'Confirm Disable',
    totp_disable_desc: 'Enter your current TOTP code to disable 2FA:',
    // Search
    tab_ebooks: 'Ebooks',
    tab_audiobooks: 'Audiobooks',
    tab_manga: 'Manga',
    search_placeholder: 'Search for books...',
    search_placeholder_ab: 'Search for audiobooks...',
    search_placeholder_manga: 'Search for manga...',
    sort_relevance: 'Relevance',
    sort_seeders: 'Seeders',
    sort_size: 'Size',
    n_results: '{n} results',
    search_empty_title: 'Search for your next read',
    search_empty_hint: 'Try searching by title, author, or ISBN',
    no_results: 'No results found',
    no_results_hint: 'Try different keywords or check your spelling',
    download: 'Download',
    download_added: 'Added',
    download_failed_state: 'Failed',
    search_failed: 'Search failed: {msg}',
    n_seeds: '{n} seed',
    n_leech: '{n} leech',
    // Search view toggles
    view_grid: 'Grid',
    view_list: 'List',
    group_flat: 'Flat',
    group_source: 'By source',
    // Swimlane status
    lane_searching: 'Searching…',
    lane_done: '{n} results',
    lane_circuit: 'Unavailable',
    lane_error: 'Error',
    // Table columns
    col_title: 'Title',
    col_author: 'Author',
    col_source: 'Source',
    col_format: 'Format',
    col_size: 'Size',
    col_seeders: 'S/L',
    col_indexer: 'Indexer',
    col_action: 'Action',
    // Library
    library_filter_placeholder: 'Filter library...',
    library_empty: 'Your library is empty',
    library_empty_hint: 'Download some books to get started',
    failed_load_library: 'Failed to load library',
    other: 'Other',
    n_items: '{n} items',
    n_files: '{n} files',
    n_pages: '{n} pages',
    open_in_abs: 'Open in Audiobookshelf',
    open_in_kavita: 'Open in Kavita',
    select_items: 'Select',
    selected: 'selected',
    edit_metadata: 'Edit metadata',
    clear: 'Clear',
    series_selected: 'series selected',
    metadata_help: 'Only fields you enable are written. Changes are saved into the book files (EPUB/CBZ) and Kavita is rescanned. Fields left disabled are untouched.',
    list_mode: 'For authors/genres/tags:',
    mode_replace: 'Replace',
    mode_append: 'Append',
    field_series: 'Series',
    field_series_hint: '(the name Kavita displays)',
    field_title: 'Title',
    field_title_hint: '(this volume/chapter)',
    field_authors: 'Authors',
    field_genres: 'Genres',
    field_tags: 'Tags',
    field_publisher: 'Publisher',
    field_language: 'Language',
    field_summary: 'Summary',
    comma_separated: '(comma-separated)',
    cancel: 'Cancel',
    save_write_files: 'Save & write files',
    metadata_no_fields: 'Enable at least one field to write.',
    rename_enable: 'Also rename on disk',
    rename_help: "Kavita names a series after its folder when the file metadata doesn't win. Renaming to the series name above fixes the display at the source. Requires the Series field.",
    rename_folder_only: 'Folder only',
    rename_folder_and_files: 'Folder and files',
    rename_needs_series: 'Enable the Series field to rename on disk.',
    rename_no_changes: 'Nothing to rename — names already match.',
    rename_previewing: 'Computing preview…',
    rename_preview_failed: 'Could not compute the rename preview.',
    writing: 'Writing…',
    metadata_online_help: 'Fetch metadata from an online database to pre-fill the fields below, then review before saving.',
    search_online: 'Search online',
    source_manga: 'Manga (AniList)',
    source_book: 'Book (Open Library)',
    searching: 'Searching…',
    no_matches: 'No matches found.',
    search_failed: 'Search failed.',
    prev: 'Prev',
    next: 'Next',
    // Downloads
    downloads_title: 'Downloads',
    refresh: 'Refresh',
    clear_completed: 'Clear Completed',
    no_active_downloads: 'No active downloads',
    no_downloads_hint: 'Search for books and click download to get started',
    failed_load_downloads: 'Failed to load downloads',
    retry: 'Retry',
    // Status
    status_downloading: 'Downloading',
    status_completed: 'Completed',
    status_error: 'Error',
    status_organizing: 'Organizing',
    status_dead_letter: 'Dead Letter',
    status_queued: 'Queued',
    status_searching: 'Searching',
    status_importing: 'Importing',
    status_retry_wait: 'Retry Wait',
    // Download actions
    download_started: 'Download started: {title}',
    download_complete: 'Download completed: {title}',
    download_completed: 'Download completed: {title}',
    download_failed: 'Download failed: {msg}',
    download_failed_anna_no_match: "Anna's Archive could not find a matching LibGen MD5 for this book. Download it manually from Anna's Archive or choose another source.",
    download_failed_anna_no_match_action: "Open Anna's Archive",
    unknown_title: 'Unknown title',
    unknown_error: 'Unknown error',
    retrying_download: 'Retrying download',
    retry_failed: 'Retry failed',
    cleared_completed: 'Cleared completed downloads',
    failed_clear: 'Failed to clear',
    // Wishlist
    wishlist_title: 'Wishlist',
    wishlist_add_book: 'Add Book',
    ph_wishlist_title: 'Title',
    ph_wishlist_author: 'Author (optional)',
    opt_ebook: 'Ebook',
    opt_audiobook: 'Audiobook',
    opt_manga: 'Manga',
    add: 'Add',
    cancel: 'Cancel',
    wishlist_empty: 'Wishlist is empty',
    wishlist_empty_hint: 'Add books you want to find later',
    wishlist_search: 'Search',
    failed_load_wishlist: 'Failed to load wishlist',
    err_title_required: 'Title is required',
    added_to_wishlist: 'Added to wishlist',
    failed_add_wishlist: 'Failed to add to wishlist',
    removed_from_wishlist: 'Removed from wishlist',
    failed_delete: 'Failed to delete',
    // Settings
    s_user_mgmt_title: 'User Management',
    add_new_user: 'Add New User',
    add_user_btn: 'Add User',
    role_user: 'User',
    role_admin: 'Admin',
    last_login: 'Last login: {date}',
    never: 'Never',
    confirm_delete_user: 'Delete user "{username}"? This cannot be undone.',
    s_connection_tests: 'Connection Tests',
    conn_prowlarr: 'Prowlarr',
    conn_qbittorrent: 'qBittorrent',
    conn_transmission: 'Transmission',
    conn_sabnzbd: 'SABnzbd',
    conn_audiobookshelf: 'Audiobookshelf',
    conn_kavita: 'Kavita',
    not_tested: 'Not tested',
    btn_test: 'Test',
    s_search_sources: 'Search Sources',
    loading_sources: 'Loading sources...',
    s_configuration: 'Configuration',
    loading: 'Loading...',
    not_configured: 'Not configured',
    no_config_data: 'No configuration data available',
    failed_load_config: 'Failed to load configuration',
    no_sources: 'No sources configured',
    enabled: 'Enabled',
    disabled: 'Disabled',
    failed_load_sources: 'Failed to load sources',
    testing: 'Testing...',
    connected: 'Connected',
    conn_error: 'Error',
    // TOTP settings toasts
    failed_setup_totp: 'Failed to setup TOTP',
    enter_6digit_code: 'Enter the 6-digit code from your app',
    totp_enabled_success: 'Two-factor authentication enabled',
    verification_failed: 'Verification failed',
    enter_totp_code: 'Enter your current TOTP code',
    totp_disabled_success: 'Two-factor authentication disabled',
    failed_disable_totp: 'Failed to disable TOTP',
    // User management toasts
    user_role_updated: 'User role updated',
    failed_update_role: 'Failed to update role',
    user_deleted: 'User deleted',
    failed_delete_user: 'Failed to delete user',
    user_created: 'User created',
    failed_create_user: 'Failed to create user',
    // Auth
    signed_in: 'Signed in successfully',
    admin_created: 'Admin account created. Welcome!',
    account_created: 'Account created. Please sign in.',
    signed_out: 'Signed out',
    err_credentials_required: 'Username and password are required',
    err_invalid_credentials: 'Invalid credentials',
    err_connection: 'Connection error',
    err_code_required: 'Code is required',
    err_invalid_code: 'Invalid code',
    backup_code_used: 'Backup code used. Consider generating new ones.',
    registration_failed: 'Registration failed',
    // Stats
    n_items_in_library: '{n} items in library',
    // Search Preferences
    search_preferences: 'Search Preferences',
    filter_non_english: 'Filter non-English results',
    filter_non_english_desc: 'When enabled, books with non-English titles are hidden from English-focused sources. Multilingual sources (Flibusta, Z-Library) always show all results.',
    filter_enabled_toast: 'Foreign language filter enabled',
    filter_disabled_toast: 'Foreign language filter disabled — showing all languages',
    filter_update_failed: 'Failed to update filter setting',
    filter_save_failed: 'Failed to save filter setting',
    // Flibusta / Z-Library config display
    flibusta_enabled: 'Flibusta',
    zlibrary_enabled: 'Z-Library',
    // Browse (filesystem library browser)
    browse_title: 'Browse',
    // Library names are translated from the root id, never from the label the
    // API returns — that label is only a fallback for an unknown id.
    browse_root_manga: 'Manga',
    browse_root_ebooks: 'Ebooks',
    browse_root_audiobooks: 'Audiobooks',
    browse_no_roots: 'No library configured',
    browse_no_roots_hint: 'Set MANGA_DIR, EBOOK_DIR or AUDIOBOOK_DIR to browse your files',
    browse_empty_dir: 'This folder is empty',
    browse_load_failed: 'Failed to load this folder',
    browse_roots_failed: 'Failed to load libraries',
    browse_parent: 'Parent folder',
    browse_folder: 'Folder',
    browse_kind_cbz: 'CBZ',
    browse_kind_epub: 'EPUB',
    browse_kind_pdf: 'PDF',
    browse_kind_cbr: 'CBR',
    browse_kind_audio: 'Audio',
    browse_kind_other: 'File',
    // Browse — manual Kavita scan (never triggered by an edit)
    browse_scan: 'Scan in Kavita',
    browse_scan_running: 'Scanning…',
    browse_scan_done: 'Kavita scan requested (all libraries)',
    browse_scan_failed: 'Failed to request the Kavita scan',
    // Browse — folder rename with preview
    browse_rename: 'Rename',
    browse_rename_title: 'Rename folder',
    browse_rename_close: 'Close',
    browse_rename_series: 'Series name',
    browse_rename_mode: 'Scope',
    browse_rename_mode_folder: 'Folder only',
    browse_rename_mode_folder_files: 'Folder and files',
    browse_rename_col_current: 'Current name',
    browse_rename_col_volume: 'Volume',
    browse_rename_col_target: 'New name',
    browse_rename_apply: 'Apply',
    browse_rename_applying: 'Applying…',
    browse_rename_folder_same: 'The folder name is already correct',
    browse_rename_folder_to: 'Folder: {from} → {to}',
    browse_rename_empty: 'This folder has no files',
    browse_rename_preview_failed: 'Failed to preview the rename',
    browse_rename_failed: 'Failed to apply the rename',
    browse_rename_done: '{count} item(s) renamed',
    browse_rename_nothing: 'Nothing to rename',
    browse_rename_summary: '{count} file(s) will be renamed',
    // Why a file is left alone — never a guess, always a stated reason.
    browse_rename_status_unchanged: 'Already named correctly',
    browse_rename_status_no_volume: 'No volume number — enter one to include it',
    browse_rename_status_collision: 'Another file already takes this name',
    browse_rename_status_folder_only: 'Not renamed in this scope',
    browse_rename_volume_placeholder: '—',
    // Single-file rename
    browse_rename_file: 'Rename this file',
    browse_rename_file_save: 'Save',
    browse_rename_file_cancel: 'Cancel',
    browse_rename_file_done: 'File renamed',
    browse_rename_file_failed: 'Failed to rename this file',
    // Metadata — folder panel
    browse_meta: 'Metadata',
    browse_meta_folder_title: 'Folder metadata',
    browse_meta_close: 'Close',
    browse_meta_folder_hint: 'Ticked fields are written to every file in this folder, replacing what is already there.',
    browse_meta_recursive: 'Include subfolders',
    browse_meta_append: 'Add to existing lists instead of replacing',
    browse_meta_preview: 'Preview',
    browse_meta_apply: 'Apply',
    browse_meta_applying: 'Applying…',
    browse_meta_save: 'Save',
    browse_meta_saving: 'Saving…',
    browse_meta_col_file: 'File',
    browse_meta_col_field: 'Field',
    browse_meta_col_current: 'Current value',
    browse_meta_col_target: 'New value',
    browse_meta_no_fields: 'Tick at least one field to edit',
    browse_meta_preview_failed: 'Failed to preview this edit',
    browse_meta_failed: 'Failed to apply this edit',
    browse_meta_load_failed: 'Failed to read this file',
    browse_meta_none: 'No file in this folder would change',
    browse_meta_summary: '{count} file(s) will be written',
    browse_meta_done: '{count} file(s) written',
    browse_meta_empty_value: '(empty)',
    // Per-file metadata panel
    browse_meta_file: 'Edit metadata',
    browse_meta_file_title: 'Metadata — {name}',
    // Field labels
    browse_meta_f_title: 'Title',
    browse_meta_f_series: 'Series',
    browse_meta_f_series_index: 'Volume number',
    browse_meta_f_summary: 'Summary',
    browse_meta_f_publisher: 'Publisher',
    browse_meta_f_language: 'Language',
    browse_meta_f_isbn: 'ISBN',
    browse_meta_f_authors: 'Authors',
    browse_meta_f_genres: 'Genres',
    browse_meta_f_tags: 'Tags',
    browse_meta_list_hint: 'comma-separated',
    // Row outcomes — keys, so an FYI never reads as a failure.
    browse_meta_status_edit: 'Will be written',
    browse_meta_status_unchanged: 'Already up to date',
    browse_meta_status_unsupported: 'Rename only',
    browse_meta_status_error: 'Could not be read',
    // Why a format cannot hold metadata — one message per format, because the
    // answer to each is different.
    browse_meta_reason_pdf_ignored_by_kavita: 'Kavita ignores PDF metadata: only the folder and file names matter, so rename instead.',
    browse_meta_reason_cbr_rar_archive: 'A CBR is a RAR archive, which librarr cannot rewrite. Convert it to CBZ to make it editable.',
    browse_meta_reason_audio_unsupported: 'Audiobook tags are outside what librarr writes; only renaming is available.',
    browse_meta_reason_format_unsupported: 'This format holds no metadata librarr can write.',
    // Informational notices, kept apart from errors.
    browse_meta_notice_truncated: 'Only the first {count} files were inspected — this preview is partial',
    browse_meta_notice_unsupported_skipped: '{count} file(s) cannot hold metadata and are left alone',
    browse_meta_notice_already_up_to_date: '{count} file(s) already hold these values',
    // Browse — online metadata lookup (folder panel only)
    browse_online_help: 'Fetch metadata from an online database to pre-fill the fields below, then preview before applying.',
    browse_online_search: 'Search online',
    browse_online_searching: 'Searching…',
    browse_online_note: 'Year and cover help you pick the right result; librarr does not write them into your files.',
    browse_online_none: 'No match found for this query.',
    // Each degraded case gets its own answer: an unconfigured source and a
    // source that failed call for different actions.
    browse_online_unavailable: 'This source is not configured on this server.',
    browse_online_upstream: 'The online source did not answer. Try again in a moment.',
    browse_online_failed: 'The search could not be run.',
    browse_online_picked: 'Fields filled from this result — review the ticked ones, then preview.',
    browse_online_unknown_title: 'Untitled',
  },
  ru: {
    // Navigation
    nav_search: 'Поиск',
    nav_library: 'Библиотека',
    nav_downloads: 'Загрузки',
    nav_wishlist: 'Желаемое',
    nav_settings: 'Настройки',
    sign_out: 'Выйти',
    // Header
    header_subtitle: 'Менеджер книг для самохостинга',
    // Login modal
    login_subtitle: 'Войдите для продолжения',
    login_sign_in: 'Войти',
    login_or: 'или',
    login_sso: 'Войти через SSO',
    login_with: 'Войти через {provider}',
    login_no_account: 'Нет аккаунта? <a href="#" data-action="showRegisterForm" class="text-indigo-400 hover:text-indigo-300">Регистрация</a>',
    create_admin_account: 'Создайте аккаунт администратора',
    create_your_account: 'Создайте аккаунт',
    create_account: 'Создать аккаунт',
    back_to_login: 'Назад к входу',
    // Labels
    label_username: 'Логин',
    label_password: 'Пароль',
    // Placeholders
    ph_username: 'Логин',
    ph_password: 'Пароль',
    ph_choose_username: 'Выберите логин',
    ph_min_6_chars: 'Минимум 6 символов',
    // TOTP
    totp_enter_code: 'Введите 6-значный код из приложения-аутентификатора или резервный код.',
    totp_code_label: 'TOTP код',
    totp_verify: 'Проверить',
    s_totp_title: 'Двухфакторная аутентификация',
    s_totp_desc: 'Добавьте дополнительный уровень защиты с помощью TOTP-приложения.',
    enable_2fa: 'Включить 2FA',
    disable_2fa: 'Отключить 2FA',
    totp_enabled_msg: 'Двухфакторная аутентификация включена.',
    totp_scan_qr: 'Отсканируйте QR-код приложением-аутентификатором (Google Authenticator, Authy и т.д.)',
    totp_manual_secret: 'Или введите секрет вручную:',
    totp_backup_codes: 'Резервные коды (сохраните в безопасном месте!):',
    verify_enable: 'Проверить и включить',
    confirm_disable: 'Подтвердить отключение',
    totp_disable_desc: 'Введите текущий TOTP-код для отключения 2FA:',
    // Search
    tab_ebooks: 'Книги',
    tab_audiobooks: 'Аудиокниги',
    tab_manga: 'Манга',
    search_placeholder: 'Поиск книг...',
    search_placeholder_ab: 'Поиск аудиокниг...',
    search_placeholder_manga: 'Поиск манги...',
    sort_relevance: 'Релевантность',
    sort_seeders: 'Сидеры',
    sort_size: 'Размер',
    n_results: '{n} результатов',
    search_empty_title: 'Найдите следующую книгу для чтения',
    search_empty_hint: 'Попробуйте искать по названию, автору или ISBN',
    no_results: 'Ничего не найдено',
    no_results_hint: 'Попробуйте другие ключевые слова или проверьте написание',
    download: 'Скачать',
    download_added: 'Добавлено',
    download_failed_state: 'Ошибка',
    search_failed: 'Ошибка поиска: {msg}',
    n_seeds: '{n} сид.',
    n_leech: '{n} лич.',
    // Search view toggles
    view_grid: 'Сетка',
    view_list: 'Список',
    group_flat: 'Списком',
    group_source: 'По источнику',
    // Swimlane status
    lane_searching: 'Поиск…',
    lane_done: '{n} результатов',
    lane_circuit: 'Недоступно',
    lane_error: 'Ошибка',
    // Table columns
    col_title: 'Название',
    col_author: 'Автор',
    col_source: 'Источник',
    col_format: 'Формат',
    col_size: 'Размер',
    col_seeders: 'С/Л',
    col_indexer: 'Индексатор',
    col_action: 'Действие',
    // Library
    library_filter_placeholder: 'Фильтр библиотеки...',
    library_empty: 'Ваша библиотека пуста',
    library_empty_hint: 'Скачайте несколько книг для начала',
    failed_load_library: 'Не удалось загрузить библиотеку',
    other: 'Другое',
    n_items: '{n} элементов',
    n_files: '{n} файлов',
    n_pages: '{n} страниц',
    open_in_abs: 'Открыть в Audiobookshelf',
    open_in_kavita: 'Открыть в Kavita',
    select_items: 'Выбрать',
    selected: 'выбрано',
    edit_metadata: 'Изменить метаданные',
    clear: 'Очистить',
    series_selected: 'серий выбрано',
    metadata_help: 'Записываются только включённые поля. Изменения сохраняются в файлы книг (EPUB/CBZ), затем Kavita пересканируется. Отключённые поля не трогаются.',
    list_mode: 'Для авторов/жанров/тегов:',
    mode_replace: 'Заменить',
    mode_append: 'Добавить',
    field_authors: 'Авторы',
    field_genres: 'Жанры',
    field_tags: 'Теги',
    field_publisher: 'Издатель',
    field_language: 'Язык',
    field_summary: 'Описание',
    comma_separated: '(через запятую)',
    cancel: 'Отмена',
    save_write_files: 'Сохранить и записать файлы',
    metadata_no_fields: 'Включите хотя бы одно поле для записи.',
    writing: 'Запись…',
    metadata_online_help: 'Получите метаданные из онлайн-базы, чтобы заполнить поля ниже, затем проверьте перед сохранением.',
    search_online: 'Искать онлайн',
    source_manga: 'Манга (AniList)',
    source_book: 'Книга (Open Library)',
    searching: 'Поиск…',
    no_matches: 'Совпадений не найдено.',
    search_failed: 'Ошибка поиска.',
    prev: 'Назад',
    next: 'Далее',
    // Downloads
    downloads_title: 'Загрузки',
    refresh: 'Обновить',
    clear_completed: 'Очистить завершённые',
    no_active_downloads: 'Нет активных загрузок',
    no_downloads_hint: 'Найдите книги и нажмите «Скачать» для начала',
    failed_load_downloads: 'Не удалось загрузить список загрузок',
    retry: 'Повторить',
    // Status
    status_downloading: 'Загрузка',
    status_completed: 'Завершено',
    status_error: 'Ошибка',
    status_organizing: 'Организация',
    status_dead_letter: 'Dead Letter',
    status_queued: 'В очереди',
    status_searching: 'Поиск',
    status_importing: 'Импорт',
    status_retry_wait: 'Ожидание повтора',
    // Download actions
    download_started: 'Загрузка начата: {title}',
    download_complete: 'Загрузка завершена: {title}',
    download_completed: 'Загрузка завершена: {title}',
    download_failed: 'Ошибка загрузки: {msg}',
    download_failed_anna_no_match: 'Anna\'s Archive не смог найти совпадающий MD5 LibGen для этой книги. Скачайте её вручную из Anna\'s Archive или выберите другой источник.',
    download_failed_anna_no_match_action: 'Открыть Anna\'s Archive',
    unknown_title: 'Неизвестное название',
    unknown_error: 'Неизвестная ошибка',
    retrying_download: 'Повтор загрузки',
    retry_failed: 'Ошибка повтора',
    cleared_completed: 'Завершённые загрузки очищены',
    failed_clear: 'Не удалось очистить',
    // Wishlist
    wishlist_title: 'Список желаемого',
    wishlist_add_book: 'Добавить книгу',
    ph_wishlist_title: 'Название',
    ph_wishlist_author: 'Автор (необязательно)',
    opt_ebook: 'Книга',
    opt_audiobook: 'Аудиокнига',
    opt_manga: 'Манга',
    add: 'Добавить',
    cancel: 'Отмена',
    wishlist_empty: 'Список желаемого пуст',
    wishlist_empty_hint: 'Добавьте книги, которые хотите найти позже',
    wishlist_search: 'Найти',
    failed_load_wishlist: 'Не удалось загрузить список желаемого',
    err_title_required: 'Требуется название',
    added_to_wishlist: 'Добавлено в список желаемого',
    failed_add_wishlist: 'Не удалось добавить в список желаемого',
    removed_from_wishlist: 'Удалено из списка желаемого',
    failed_delete: 'Не удалось удалить',
    // Settings
    s_user_mgmt_title: 'Управление пользователями',
    add_new_user: 'Добавить пользователя',
    add_user_btn: 'Добавить',
    role_user: 'Пользователь',
    role_admin: 'Админ',
    last_login: 'Последний вход: {date}',
    never: 'Никогда',
    confirm_delete_user: 'Удалить пользователя «{username}»? Это нельзя отменить.',
    s_connection_tests: 'Проверка подключений',
    conn_prowlarr: 'Prowlarr',
    conn_qbittorrent: 'qBittorrent',
    conn_transmission: 'Transmission',
    conn_sabnzbd: 'SABnzbd',
    conn_audiobookshelf: 'Audiobookshelf',
    conn_kavita: 'Kavita',
    not_tested: 'Не проверено',
    btn_test: 'Проверить',
    s_search_sources: 'Источники поиска',
    loading_sources: 'Загрузка источников...',
    s_configuration: 'Конфигурация',
    loading: 'Загрузка...',
    not_configured: 'Не настроено',
    no_config_data: 'Нет данных конфигурации',
    failed_load_config: 'Не удалось загрузить конфигурацию',
    no_sources: 'Нет настроенных источников',
    enabled: 'Включено',
    disabled: 'Отключено',
    failed_load_sources: 'Не удалось загрузить источники',
    testing: 'Проверка...',
    connected: 'Подключено',
    conn_error: 'Ошибка',
    // TOTP settings toasts
    failed_setup_totp: 'Не удалось настроить TOTP',
    enter_6digit_code: 'Введите 6-значный код из приложения',
    totp_enabled_success: 'Двухфакторная аутентификация включена',
    verification_failed: 'Ошибка проверки',
    enter_totp_code: 'Введите текущий TOTP-код',
    totp_disabled_success: 'Двухфакторная аутентификация отключена',
    failed_disable_totp: 'Не удалось отключить TOTP',
    // User management toasts
    user_role_updated: 'Роль пользователя обновлена',
    failed_update_role: 'Не удалось обновить роль',
    user_deleted: 'Пользователь удалён',
    failed_delete_user: 'Не удалось удалить пользователя',
    user_created: 'Пользователь создан',
    failed_create_user: 'Не удалось создать пользователя',
    // Auth
    signed_in: 'Успешный вход',
    admin_created: 'Аккаунт администратора создан. Добро пожаловать!',
    account_created: 'Аккаунт создан. Пожалуйста, войдите.',
    signed_out: 'Вы вышли',
    err_credentials_required: 'Требуется логин и пароль',
    err_invalid_credentials: 'Неверные учётные данные',
    err_connection: 'Ошибка соединения',
    err_code_required: 'Требуется код',
    err_invalid_code: 'Неверный код',
    backup_code_used: 'Использован резервный код. Рекомендуется создать новые.',
    registration_failed: 'Ошибка регистрации',
    // Stats
    n_items_in_library: '{n} элементов в библиотеке',
    // Search Preferences
    search_preferences: 'Настройки поиска',
    filter_non_english: 'Фильтровать неанглоязычные результаты',
    filter_non_english_desc: 'Если включено, книги с неанглоязычными названиями скрываются из англоязычных источников. Многоязычные источники (Flibusta, Z-Library) всегда показывают все результаты.',
    filter_enabled_toast: 'Фильтр иностранных языков включён',
    filter_disabled_toast: 'Фильтр иностранных языков выключен — показаны все языки',
    filter_update_failed: 'Не удалось обновить настройку фильтра',
    filter_save_failed: 'Не удалось сохранить настройку фильтра',
    // Flibusta / Z-Library config display
    flibusta_enabled: 'Flibusta',
    zlibrary_enabled: 'Z-Library',
  }
};

// ============================================================
// I18N FUNCTIONS
// ============================================================
function t(key, vars) {
  const lang = I18N[currentLang] || I18N.en;
  let val = lang[key] || I18N.en[key] || key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      val = val.replaceAll(`{${k}}`, v);
    }
  }
  return val;
}

function applyLanguage() {
  document.documentElement.lang = currentLang;
  document.getElementById('lang-toggle').textContent = currentLang === 'en' ? 'RU' : 'EN';

  document.querySelectorAll('[data-i18n]').forEach(el => {
    el.textContent = t(el.dataset.i18n);
  });
  document.querySelectorAll('[data-i18n-html]').forEach(el => {
    el.innerHTML = t(el.dataset.i18nHtml);
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    el.placeholder = t(el.dataset.i18nPlaceholder);
  });
  document.querySelectorAll('[data-i18n-title]').forEach(el => {
    el.title = t(el.dataset.i18nTitle);
  });
}

function toggleLanguage() {
  currentLang = currentLang === 'en' ? 'ru' : 'en';
  localStorage.setItem('librarr_lang', currentLang);
  applyLanguage();
  refreshDynamicContent();
}

function refreshDynamicContent() {
  // Re-render current tab content with new language
  const tab = state.currentTab;
  if (tab === 'search' && state.searchResults.length > 0) {
    renderSearchResults();
  } else if (tab === 'downloads') {
    refreshDownloads();
  } else if (tab === 'library') {
    loadLibrary();
  } else if (tab === 'browse') {
    loadBrowse();
  } else if (tab === 'wishlist') {
    loadWishlist();
  } else if (tab === 'settings') {
    loadConfig();
    loadSources();
  }
}

// ============================================================
// STATE
// ============================================================
const state = {
  currentTab: 'search',
  searchTab: 'ebooks',
  libraryTab: 'ebooks',
  searchResults: [],
  pendingDownloads: new Set(),
  trackedDownloadJobs: new Map(),
  downloadOutcomes: new Map(),
  downloadOutcomeTimers: new Map(),
  downloadJobs: [],
  pendingRetryDownloads: new Set(),
  sortMode: 'relevance',
  sortDir: 'desc',
  // Search view preferences (persisted). Default: list layout + swimlanes.
  viewLayout: localStorage.getItem('librarr.search.viewLayout') || 'list', // 'grid' | 'list'
  viewGroup: localStorage.getItem('librarr.search.viewGroup') || 'source', // 'flat' | 'source'
  collapsedLanes: new Set(),
  sources: [],                 // filtered SourceMeta for the current tab (lane order)
  sourceStatus: new Map(),     // source name -> { status, count, timeMs, error }
  resultsByKey: new Map(),     // stable result key -> result (replaces data-idx)
  libraryPage: 1,
  libraryPages: 1,
  // Browse tab (filesystem browser, /api/files/*).
  browseRoots: [],
  browseRootId: null,
  browsePath: null,
  browseIsRoot: true,
  browseKavita: false,
  // Rename panel: the folder it was opened on, and the volumes the user typed
  // (file name -> number). Kept out of the DOM so a re-render never loses them.
  renamePath: null,
  renameVolumes: {},
  renamePreview: null,
  // Metadata panels: the folder / file each was opened on. metaPreview holds the
  // last preview, and is dropped as soon as any input changes — applying a plan
  // the user is no longer looking at is exactly what decision 4 forbids.
  metaPath: null,
  metaPreview: null,
  metaFilePath: null,
  metaFileEditable: false,
  // Bulk metadata selection (manga/Kavita tab only).
  selectMode: false,
  selectedSeries: new Map(), // Kavita series id -> name
  config: null,
  downloadPollTimer: null,
  currentUser: null,
  currentRole: null,
};

// Some sources emit a result.source that differs from their SourceMeta name
// (e.g. Prowlarr ebooks → "torrent", AudioBookBay → "audiobook"). Map the
// result source onto the lane/meta name so swimlanes reconcile correctly.
const RESULT_SOURCE_ALIASES = {
  torrent: 'prowlarr',
  audiobook: 'audiobookbay',
};
function laneKeyFor(source) {
  return RESULT_SOURCE_ALIASES[source] || source || 'unknown';
}

// Maps the frontend search tab onto the backend SearchTab() value used in
// SourceMeta so we can filter sources to the current tab.
const SEARCH_TAB_META = { ebooks: 'main', audiobooks: 'audiobook', manga: 'manga' };

const SOURCE_COLORS = {
  annas:           { bg: '#7c3aed', text: 'white', label: "Anna's Archive" },
  torrent:         { bg: '#2563eb', text: 'white', label: 'Prowlarr' },
  prowlarr_manga:  { bg: '#2563eb', text: 'white', label: 'Prowlarr' },
  audiobook:       { bg: '#2563eb', text: 'white', label: 'Prowlarr' },
  audiobookbay:    { bg: '#059669', text: 'white', label: 'AudioBookBay' },
  gutenberg:       { bg: '#d97706', text: 'white', label: 'Gutenberg' },
  openlibrary:     { bg: '#dc2626', text: 'white', label: 'Open Library' },
  standardebooks:  { bg: '#0891b2', text: 'white', label: 'Standard Ebooks' },
  librivox:        { bg: '#7c3aed', text: 'white', label: 'Librivox' },
  mangadex:        { bg: '#ff6740', text: 'white', label: 'MangaDex' },
  nyaa_manga:      { bg: '#16a34a', text: 'white', label: 'Nyaa' },
  annas_manga:     { bg: '#7c3aed', text: 'white', label: "Anna's Manga" },
  webnovel:        { bg: '#6366f1', text: 'white', label: 'Web Novel' },
  flibusta:        { bg: '#b91c1c', text: 'white', label: 'Flibusta' },
  zlibrary:        { bg: '#4338ca', text: 'white', label: 'Z-Library' },
  tpb:             { bg: '#1e40af', text: 'white', label: 'ThePirateBay' },
  tpb_audiobook:   { bg: '#1e40af', text: 'white', label: 'ThePirateBay' },
  booktracker:     { bg: '#0e7490', text: 'white', label: 'BookTracker' },
  booktracker_audiobook: { bg: '#0e7490', text: 'white', label: 'BookTracker' },
};

const COVER_GRADIENTS = [
  'from-indigo-600 to-purple-700',
  'from-blue-600 to-cyan-700',
  'from-emerald-600 to-teal-700',
  'from-rose-600 to-pink-700',
  'from-amber-600 to-orange-700',
  'from-violet-600 to-fuchsia-700',
  'from-sky-600 to-blue-700',
  'from-lime-600 to-green-700',
];

const STATUS_STYLES = {
  downloading: { bg: 'bg-blue-500/20', text: 'text-blue-400', border: 'border-blue-500/30', label: 'Downloading' },
  completed:   { bg: 'bg-emerald-500/20', text: 'text-emerald-400', border: 'border-emerald-500/30', label: 'Completed' },
  error:       { bg: 'bg-red-500/20', text: 'text-red-400', border: 'border-red-500/30', label: 'Error' },
  organizing:  { bg: 'bg-yellow-500/20', text: 'text-yellow-400', border: 'border-yellow-500/30', label: 'Organizing' },
  dead_letter: { bg: 'bg-slate-500/20', text: 'text-slate-400', border: 'border-slate-500/30', label: 'Dead Letter' },
  queued:      { bg: 'bg-slate-500/20', text: 'text-slate-400', border: 'border-slate-500/30', label: 'Queued' },
  searching:   { bg: 'bg-indigo-500/20', text: 'text-indigo-400', border: 'border-indigo-500/30', label: 'Searching' },
  importing:   { bg: 'bg-purple-500/20', text: 'text-purple-400', border: 'border-purple-500/30', label: 'Importing' },
  retry_wait:  { bg: 'bg-amber-500/20', text: 'text-amber-400', border: 'border-amber-500/30', label: 'Retry Wait' },
};

const TERMINAL_DOWNLOAD_STATUSES = new Set(['completed', 'error', 'dead_letter']);

// ============================================================
// API HELPERS
// ============================================================
function getApiKey() {
  return localStorage.getItem('librarr_apikey') || '';
}

async function api(path, options = {}) {
  const url = new URL(path, window.location.origin);
  const key = getApiKey();
  if (key) url.searchParams.set('apikey', key);

  const resp = await fetch(url.toString(), {
    credentials: 'include',
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  });

  if (resp.status === 401) {
    showLoginModal();
    throw new Error('Unauthorized');
  }

  return resp;
}

async function apiJson(path, options = {}) {
  const resp = await api(path, options);
  if (!resp.ok) throw new Error(`API error: ${resp.status}`);
  return resp.json();
}

// ============================================================
// TOAST NOTIFICATIONS
// ============================================================
function showToast(message, type = 'info', options = {}) {
  const container = document.getElementById('toast-container');
  const {
    sticky = false,
    actionLabel = '',
    actionHref = '',
    actionTarget = '_blank',
  } = options;
  const colors = {
    info: 'bg-slate-800 border-slate-600',
    success: 'bg-emerald-900/80 border-emerald-600/50',
    error: 'bg-red-900/80 border-red-600/50',
    warning: 'bg-yellow-900/80 border-yellow-600/50',
  };
  const icons = {
    info: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>',
    success: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>',
    error: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"/>',
    warning: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>',
  };
  const iconColors = { info: 'text-slate-400', success: 'text-emerald-400', error: 'text-red-400', warning: 'text-yellow-400' };

  const el = document.createElement('div');
  el.className = `toast-enter ${colors[type]} border rounded-lg px-4 py-3 shadow-xl flex items-start gap-3`;
  const messageWrap = document.createElement('div');
  messageWrap.className = 'flex-1 min-w-0';

  const messageEl = document.createElement('p');
  messageEl.className = 'text-sm text-slate-200';
  messageEl.textContent = message;
  messageWrap.appendChild(messageEl);

  if (actionLabel && actionHref) {
    const action = document.createElement('a');
    action.href = actionHref;
    action.target = actionTarget;
    action.rel = 'noreferrer noopener';
    action.className = 'mt-2 inline-flex items-center rounded-lg bg-indigo-600 px-2.5 py-1 text-xs font-medium text-white transition-colors hover:bg-indigo-500';
    action.textContent = actionLabel;
    messageWrap.appendChild(action);
  }

  const close = document.createElement('button');
  close.className = 'text-slate-500 hover:text-slate-300 flex-shrink-0';
  close.addEventListener('click', () => el.remove());
  close.innerHTML = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>';

  const icon = document.createElement('svg');
  icon.className = `w-5 h-5 ${iconColors[type]} flex-shrink-0 mt-0.5`;
  icon.setAttribute('fill', 'none');
  icon.setAttribute('stroke', 'currentColor');
  icon.setAttribute('viewBox', '0 0 24 24');
  icon.innerHTML = icons[type];

  el.appendChild(icon);
  el.appendChild(messageWrap);
  el.appendChild(close);
  container.appendChild(el);

  if (!sticky) {
    setTimeout(() => {
      el.classList.remove('toast-enter');
      el.classList.add('toast-exit');
      setTimeout(() => el.remove(), 300);
    }, 4000);
  }
}

function isAnnaNoMatchError(msg) {
  return typeof msg === 'string' && (
    msg.includes('matching LibGen MD5') ||
    msg.includes('libgen no matching MD5') ||
    msg.includes('File not found in DB')
  );
}

function showAnnaNoMatchToast(title, annaUrl) {
  showToast(
    t('download_failed_anna_no_match'),
    'error',
    annaUrl ? {
      sticky: true,
      actionLabel: t('download_failed_anna_no_match_action'),
      actionHref: annaUrl,
    } : {
      sticky: true,
    }
  );
}

// ============================================================
// LOGIN / REGISTER / TOTP
// ============================================================
let pendingTOTPSession = '';

function showLoginModal() {
  document.getElementById('login-modal').classList.remove('hidden');
  document.getElementById('login-modal').classList.add('flex');
  showLoginForm();

  // Check auth status to decide showing register link or OIDC button.
  fetch('/api/auth/status', { credentials: 'include' }).then(r => r.json()).then(data => {
    if (data.authenticated) {
      hideLoginModal();
      return;
    }
    // Always show register link — invite codes make self-registration secure.
    // First user creates admin (no invite needed). After that, invite code required.
    const regLink = document.getElementById('login-register-link');
    regLink.classList.remove('hidden');
    if (!data.has_users) {
      // First-run: default to the register form with a welcome banner. The
      // login form has nothing to log into yet, so showing it first wastes a
      // click and hides the actual setup step behind a "Register" link.
      document.getElementById('login-subtitle').textContent = t('create_admin_account');
      document.getElementById('first-run-banner').classList.remove('hidden');
      const invField = document.getElementById('invite-code-field');
      if (invField) invField.classList.add('hidden');
      showRegisterForm();
      // No accounts exist — hide the "Back to login" link in register form.
      const backLinks = document.querySelectorAll('#register-form a[data-action="showLoginForm"]');
      backLinks.forEach(a => a.parentElement.classList.add('hidden'));
    } else {
      document.getElementById('login-subtitle').textContent = t('login_subtitle');
      document.getElementById('first-run-banner').classList.add('hidden');
    }
  }).catch(() => {});

  // Check OIDC config.
  fetch('/api/auth/status', { credentials: 'include' }).then(r => r.json()).then(data => {
    if (data.oidc_enabled) {
      document.getElementById('login-oidc-btn').classList.remove('hidden');
      document.getElementById('oidc-login-link').textContent = t('login_with', {provider: data.oidc_provider_name || 'SSO'});
    }
  }).catch(() => {});
}

function hideLoginModal() {
  document.getElementById('login-modal').classList.add('hidden');
  document.getElementById('login-modal').classList.remove('flex');
}

function showLoginForm() {
  document.getElementById('login-form').classList.remove('hidden');
  document.getElementById('totp-form').classList.add('hidden');
  document.getElementById('register-form').classList.add('hidden');
  document.getElementById('login-username').focus();
}

function showRegisterForm() {
  document.getElementById('login-form').classList.add('hidden');
  document.getElementById('totp-form').classList.add('hidden');
  document.getElementById('register-form').classList.remove('hidden');
  document.getElementById('login-subtitle').textContent = t('create_your_account');
  document.getElementById('register-username').focus();
}

function showTOTPForm() {
  document.getElementById('login-form').classList.add('hidden');
  document.getElementById('register-form').classList.add('hidden');
  document.getElementById('totp-form').classList.remove('hidden');
  document.getElementById('login-subtitle').textContent = t('s_totp_title');
  document.getElementById('totp-code').focus();
}

document.getElementById('login-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const errEl = document.getElementById('login-error');
  errEl.classList.add('hidden');

  const username = document.getElementById('login-username').value.trim();
  const password = document.getElementById('login-password').value;

  if (!username || !password) {
    errEl.textContent = t('err_credentials_required');
    errEl.classList.remove('hidden');
    return;
  }

  try {
    const resp = await fetch('/api/login', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    const data = await resp.json().catch(() => ({}));

    if (resp.ok && data.success) {
      if (data.needs_totp) {
        pendingTOTPSession = data.session_pending;
        showTOTPForm();
        return;
      }
      hideLoginModal();
      updateUserHeader(data.username, data.role);
      init();
      showToast(t('signed_in'), 'success');
    } else {
      errEl.textContent = data.error || t('err_invalid_credentials');
      errEl.classList.remove('hidden');
    }
  } catch (err) {
    errEl.textContent = t('err_connection');
    errEl.classList.remove('hidden');
  }
});

document.getElementById('totp-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const errEl = document.getElementById('totp-error');
  errEl.classList.add('hidden');

  const code = document.getElementById('totp-code').value.trim();
  if (!code) {
    errEl.textContent = t('err_code_required');
    errEl.classList.remove('hidden');
    return;
  }

  try {
    const resp = await fetch('/api/login/totp', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ session_pending: pendingTOTPSession, code }),
    });

    const data = await resp.json().catch(() => ({}));
    if (resp.ok && data.success) {
      hideLoginModal();
      updateUserHeader(data.username, data.role);
      init();
      showToast(t('signed_in'), 'success');
      if (data.backup_code_used) showToast(t('backup_code_used'), 'warning');
    } else {
      errEl.textContent = data.error || t('err_invalid_code');
      errEl.classList.remove('hidden');
    }
  } catch (err) {
    errEl.textContent = t('err_connection');
    errEl.classList.remove('hidden');
  }
});

document.getElementById('register-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const errEl = document.getElementById('register-error');
  errEl.classList.add('hidden');

  const username = document.getElementById('register-username').value.trim();
  const password = document.getElementById('register-password').value;
  const inviteCode = (document.getElementById('register-invite-code')?.value || '').trim();

  if (!username || !password) {
    errEl.textContent = t('err_credentials_required');
    errEl.classList.remove('hidden');
    return;
  }

  try {
    const body = { username, password };
    if (inviteCode) body.invite_code = inviteCode;

    const resp = await fetch('/api/register', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });

    const data = await resp.json().catch(() => ({}));
    if (resp.ok && data.success) {
      if (data.token) {
        // First user — auto-logged in.
        hideLoginModal();
        updateUserHeader(data.username, data.role);
        init();
        showToast(t('admin_created'), 'success');
      } else {
        showLoginForm();
        showToast(t('account_created'), 'success');
      }
    } else {
      errEl.textContent = data.error || t('registration_failed');
      errEl.classList.remove('hidden');
    }
  } catch (err) {
    errEl.textContent = t('err_connection');
    errEl.classList.remove('hidden');
  }
});

function updateUserHeader(username, role) {
  if (username) {
    document.getElementById('header-user').classList.remove('hidden');
    document.getElementById('header-user').classList.add('flex');
    document.getElementById('header-username').textContent = username;
    document.getElementById('header-role').textContent = role || 'user';
    document.getElementById('logout-btn').classList.remove('hidden');
    state.currentUser = username;
    state.currentRole = role;
  } else {
    document.getElementById('header-user').classList.add('hidden');
    document.getElementById('logout-btn').classList.add('hidden');
    state.currentUser = null;
    state.currentRole = null;
  }
}

async function doLogout() {
  try {
    await fetch('/api/logout', { method: 'POST', credentials: 'include' });
  } catch (e) {}
  updateUserHeader(null, null);
  showLoginModal();
  showToast(t('signed_out'), 'info');
}

// ============================================================
// MOBILE NAV
// ============================================================
function toggleMobileNav() {
  const nav = document.getElementById('main-nav');
  nav.classList.toggle('mobile-open');
}

// Close mobile nav when a tab is selected
function closeMobileNav() {
  const nav = document.getElementById('main-nav');
  nav.classList.remove('mobile-open');
}

// TAB NAVIGATION
// ============================================================
function switchTab(tab) {
  closeMobileNav();
  state.currentTab = tab;

  // Update nav
  document.querySelectorAll('.nav-tab').forEach(el => {
    el.classList.toggle('active', el.dataset.tab === tab);
  });

  // Update content
  document.querySelectorAll('.tab-content').forEach(el => {
    el.classList.toggle('active', el.id === `tab-${tab}`);
  });

  // Load data for the tab
  if (tab === 'library') loadLibrary();
  if (tab === 'browse') loadBrowse();
  if (tab === 'downloads') { refreshDownloads(); startDownloadPolling(); }
  else stopDownloadPolling();
  if (tab === 'wishlist') loadWishlist();
  if (tab === 'settings') loadSettings();
}

function switchSearchTab(tab) {
  state.searchTab = tab;
  document.querySelectorAll('.sub-tab').forEach(el => {
    el.classList.toggle('active', el.dataset.stab === tab);
  });
  // Clear results when switching
  document.getElementById('search-results').innerHTML = '';
  document.getElementById('search-toolbar').classList.add('hidden');
  document.getElementById('search-no-results').classList.add('hidden');
  document.getElementById('search-empty').classList.remove('hidden');
  hideSearchSkeleton();
  state.searchResults = [];
  state.sources = [];
  state.sourceStatus = new Map();
  state.resultsByKey = new Map();
  state.collapsedLanes = new Set();

  // Update placeholder
  const placeholders = { ebooks: t('search_placeholder'), audiobooks: t('search_placeholder_ab'), manga: t('search_placeholder_manga') };
  document.getElementById('search-input').placeholder = placeholders[tab] || 'Search...';
  document.getElementById('search-input').value = '';
}

function switchLibraryTab(tab) {
  state.libraryTab = tab;
  state.libraryPage = 1;
  document.querySelectorAll('.lib-tab').forEach(el => {
    el.classList.toggle('active', el.dataset.ltab === tab);
  });
  // Bulk metadata editing is backed by Kavita series, which only the manga
  // tab lists. Leaving the tab exits select mode so state can't leak across.
  const selectToggle = document.getElementById('library-select-toggle');
  if (tab === 'manga') {
    selectToggle.classList.remove('hidden');
  } else {
    selectToggle.classList.add('hidden');
    exitSelectMode();
  }
  loadLibrary();
}

// ============================================================
// SEARCH
// ============================================================
let searchTimeout = null;
let searchAbort = null;   // AbortController for in-flight request
let searchGeneration = 0; // monotonic counter — stale responses are discarded

document.getElementById('search-input').addEventListener('input', (e) => {
  clearTimeout(searchTimeout);
  const q = e.target.value.trim();
  if (q.length < 2) return;
  searchTimeout = setTimeout(() => doSearch(q), 300);
});

document.getElementById('search-input').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') {
    clearTimeout(searchTimeout);
    const q = e.target.value.trim();
    if (q.length >= 1) doSearch(q);
  }
});

async function doSearch(query) {
  const endpoints = { ebooks: '/api/search', audiobooks: '/api/search/audiobooks', manga: '/api/search/manga' };
  const endpoint = endpoints[state.searchTab] || '/api/search';
  const streamEndpoint = `${endpoint}/stream`;

  // Abort any in-flight search — prevents stale results from overwriting
  // the new search when the old request finishes after the new one starts.
  if (searchAbort) searchAbort.abort();
  searchAbort = new AbortController();
  const gen = ++searchGeneration;

  // Reset search state for the new query.
  state.searchResults = [];
  state.sources = [];
  state.sourceStatus = new Map();
  state.resultsByKey = new Map();
  state.collapsedLanes = new Set();
  document.getElementById('search-results').innerHTML = '';
  document.getElementById('search-empty').classList.add('hidden');
  document.getElementById('search-no-results').classList.add('hidden');
  document.getElementById('search-spinner').classList.remove('hidden');
  showSearchSkeleton();

  try {
    await doStreamingSearch(streamEndpoint, query, gen, searchAbort.signal);
  } catch (err) {
    if (err.name === 'AbortError') return; // expected — new search superseded this one
    if (gen !== searchGeneration) return;
    try {
      await doJsonSearch(endpoint, query, gen, searchAbort.signal);
    } catch (fallbackErr) {
      if (fallbackErr.name === 'AbortError') return;
      if (gen !== searchGeneration) return;
      document.getElementById('search-spinner').classList.add('hidden');
      hideSearchSkeleton();
      if (fallbackErr.message !== 'Unauthorized') {
        showToast(t('search_failed', {msg: fallbackErr.message}), 'error');
      }
    }
  }
}

async function doJsonSearch(endpoint, query, gen, signal) {
  const data = await apiJson(`${endpoint}?q=${encodeURIComponent(query)}`, { signal });
  if (gen !== searchGeneration) return;
  // Degraded mode: no per-source metadata. Derive lanes from the results
  // themselves so the swimlane view still works.
  state.sources = deriveSourcesFromResults(data.results || []);
  state.sourceStatus = new Map(state.sources.map(s => [s.name, { status: 'done', count: 0, timeMs: null }]));
  applySearchResults(data.results || [], false);
}

// Filters SourceMeta to the sources relevant to the current search tab.
function filterSourceMeta(meta) {
  const wantTab = SEARCH_TAB_META[state.searchTab] || 'main';
  return (meta || []).filter(s => s && s.enabled && s.search_tab === wantTab);
}

// Builds pseudo-source lanes from a flat result list (JSON fallback).
function deriveSourcesFromResults(results) {
  const seen = new Set();
  const out = [];
  for (const r of results || []) {
    const key = laneKeyFor(r.source);
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ name: key, label: laneLabel(key), enabled: true });
  }
  return out;
}

async function doStreamingSearch(endpoint, query, gen, signal) {
  const resp = await api(`${endpoint}?q=${encodeURIComponent(query)}`, {
    signal,
    headers: { Accept: 'text/event-stream' },
  });
  if (!resp.ok || !resp.body) throw new Error(`API error: ${resp.status}`);

  const reader = resp.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let completed = false;

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const frames = buffer.split('\n\n');
    buffer = frames.pop() || '';
    for (const frame of frames) {
      const evt = parseSSEFrame(frame);
      if (!evt || gen !== searchGeneration) continue;
      if (evt.event === 'started') {
        handleSearchStarted(evt.data);
      } else if (evt.event === 'results') {
        handleSearchResults(evt.data);
      } else if (evt.event === 'complete') {
        handleSearchComplete(evt.data);
        completed = true;
      }
    }
  }

  if (gen === searchGeneration && !completed) {
    // Stream ended without a `complete` event — settle any still-searching lanes.
    handleSearchComplete({ results: state.searchResults });
  }
}

// SSE `started`: seed the lane skeleton from SourceMeta.
function handleSearchStarted(data) {
  state.sources = filterSourceMeta(data.sources);
  state.sourceStatus = new Map();
  for (const s of state.sources) {
    state.sourceStatus.set(s.name, { status: 'searching', count: 0, timeMs: null, error: '' });
  }
  applySearchResults(data.results || [], true, { rebuild: true });
}

// SSE `results`: one source responded. Update its lane status + counts.
function handleSearchResults(data) {
  // A frame may reference sources not in the initial meta (rare); make sure
  // every referenced source has a lane so nothing is dropped.
  ensureSource(data.source);
  const status = state.sourceStatus.get(laneKeyFor(data.source)) || state.sourceStatus.get(data.source);
  if (status) {
    status.timeMs = data.search_time_ms ?? status.timeMs;
    if (data.error) { status.status = 'error'; status.error = data.error; }
    else if (data.circuit_open) status.status = 'circuit';
    else if (data.source_done) status.status = 'done';
    else status.status = 'searching';
  }
  applySearchResults(data.results || [], true, { changedSource: data.source });
}

// SSE `complete`: mark every remaining lane done and do a final full render.
function handleSearchComplete(data) {
  for (const status of state.sourceStatus.values()) {
    if (status.status === 'searching') status.status = 'done';
  }
  applySearchResults(data.results || [], false, { rebuild: true });
}

// Ensures a lane/status entry exists for a source seen mid-stream.
function ensureSource(source) {
  if (!source) return;
  const key = laneKeyFor(source);
  if (state.sourceStatus.has(key) || state.sourceStatus.has(source)) return;
  state.sources.push({ name: key, label: laneLabel(key), enabled: true });
  state.sourceStatus.set(key, { status: 'searching', count: 0, timeMs: null, error: '' });
}

function parseSSEFrame(frame) {
  let event = 'message';
  const dataLines = [];
  frame.split('\n').forEach(line => {
    if (line.startsWith('event:')) event = line.slice(6).trim();
    if (line.startsWith('data:')) dataLines.push(line.slice(5).trimStart());
  });
  if (dataLines.length === 0) return null;
  try {
    return { event, data: JSON.parse(dataLines.join('\n')) };
  } catch {
    return null;
  }
}

// Central entry point for a new batch of results (streaming frame, complete,
// or JSON fallback). Rebuilds the keyed index and renders — either fully or,
// during streaming swimlane mode, just the lane that changed.
function applySearchResults(results, searching, opts = {}) {
  state.searchResults = results || [];
  buildResultsByKey();

  const hasLanes = !!document.querySelector('#search-results .search-lane');
  const canTargetLane = opts.changedSource && !opts.rebuild &&
    state.viewGroup === 'source' && hasLanes;

  if (canTargetLane) {
    applyLaneFrame(opts.changedSource);
    refreshLaneCounts();
  } else {
    renderSearchResults();
  }
  updateSearchChrome(searching);
}

// Builds state.resultsByKey (stable key -> result) and stamps result.__key,
// replacing the old data-idx scheme. Collisions get a numeric suffix so every
// clickable element resolves to exactly one result.
function buildResultsByKey() {
  const map = new Map();
  for (const r of state.searchResults) {
    let key = getDownloadKey(r);
    if (map.has(key)) {
      let n = 2;
      while (map.has(`${key}#${n}`)) n++;
      key = `${key}#${n}`;
    }
    r.__key = key;
    map.set(key, r);
  }
  state.resultsByKey = map;
}

function showSearchSkeleton() {
  const container = document.getElementById('search-skeleton');
  container.innerHTML = '';
  for (let i = 0; i < 4; i++) {
    container.innerHTML += `
      <div class="bg-slate-900 rounded-xl border border-slate-800 p-3 space-y-2">
        <div class="skeleton h-4 w-1/3 rounded"></div>
        <div class="skeleton h-3 w-2/3 rounded"></div>
        <div class="skeleton h-3 w-1/2 rounded"></div>
      </div>
    `;
  }
  container.classList.remove('hidden');
}

function hideSearchSkeleton() {
  document.getElementById('search-skeleton').classList.add('hidden');
}

function updateSearchChrome(searching) {
  const statuses = [...state.sourceStatus.values()];
  const anySearching = searching && statuses.some(s => s.status === 'searching');
  document.getElementById('search-spinner').classList.toggle('hidden', !anySearching);

  const total = state.searchResults.length;
  const started = state.sources.length > 0 || total > 0;
  document.getElementById('search-toolbar').classList.toggle('hidden', !started);
  document.getElementById('search-result-count').textContent = t('n_results', {n: total});
  if (started) document.getElementById('search-empty').classList.add('hidden');

  const settled = !anySearching && !statuses.some(s => s.status === 'searching');
  const noResults = settled && total === 0;
  document.getElementById('search-no-results').classList.toggle('hidden', !noResults);
}

// Toggles sort mode / direction and re-renders. Numeric columns default to
// descending; text columns to ascending. Clicking the active column flips it.
function setSortMode(mode) {
  const numeric = mode === 'seeders' || mode === 'size';
  if (state.sortMode === mode) {
    state.sortDir = state.sortDir === 'asc' ? 'desc' : 'asc';
  } else {
    state.sortMode = mode;
    state.sortDir = numeric ? 'desc' : 'asc';
  }
  renderSearchResults();
}

function sortResults(results) {
  const sorted = [...results];
  const mode = state.sortMode;
  if (mode === 'relevance' || !mode) return sorted; // preserve server order
  const dir = state.sortDir === 'asc' ? 1 : -1;
  const comparators = {
    seeders: (a, b) => (a.seeders || 0) - (b.seeders || 0),
    size: (a, b) => resultSizeBytes(a) - resultSizeBytes(b),
    title: (a, b) => (a.title || '').localeCompare(b.title || ''),
    author: (a, b) => (a.author || '').localeCompare(b.author || ''),
    source: (a, b) => laneLabel(laneKeyFor(a.source)).localeCompare(laneLabel(laneKeyFor(b.source))),
    format: (a, b) => (a.format || '').localeCompare(b.format || ''),
    indexer: (a, b) => (a.indexer || '').localeCompare(b.indexer || ''),
  };
  const cmp = comparators[mode];
  if (cmp) sorted.sort((a, b) => cmp(a, b) * dir);
  return sorted;
}

function parseSize(sizeStr) {
  if (!sizeStr) return 0;
  const s = sizeStr.toString().toUpperCase();
  const match = s.match(/([\d.]+)\s*(GB|MB|KB|B)?/);
  if (!match) return 0;
  const num = parseFloat(match[1]);
  const unit = match[2] || 'B';
  const multipliers = { B: 1, KB: 1024, MB: 1048576, GB: 1073741824 };
  return num * (multipliers[unit] || 1);
}

function resultSizeBytes(r) {
  if (typeof r.size === 'number' && r.size > 0) return r.size;
  return parseSize(r.size_human || r.sizeHuman || r.size || '');
}

function resultSizeText(r) {
  return (r.size && r.size > 0) ? formatSize(r.size) : (r.size_human || r.sizeHuman || '');
}

// ---- View preference setters (persisted) ----
function setViewLayout(layout) {
  if (layout !== 'grid' && layout !== 'list') return;
  state.viewLayout = layout;
  localStorage.setItem('librarr.search.viewLayout', layout);
  renderSearchResults();
}

function setViewGroup(group) {
  if (group !== 'flat' && group !== 'source') return;
  state.viewGroup = group;
  localStorage.setItem('librarr.search.viewGroup', group);
  renderSearchResults();
}

function toggleLane(source) {
  if (state.collapsedLanes.has(source)) state.collapsedLanes.delete(source);
  else state.collapsedLanes.add(source);
  const lane = document.querySelector(`#search-results .search-lane[data-lane="${cssEscape(source)}"]`);
  if (lane) {
    const body = lane.querySelector('[data-lane-body]');
    const chevron = lane.querySelector('[data-lane-chevron]');
    const collapsed = state.collapsedLanes.has(source);
    if (body) body.classList.toggle('hidden', collapsed);
    if (chevron) chevron.classList.toggle('rotate-180', collapsed);
  }
}

function cssEscape(s) {
  return String(s).replace(/["\\]/g, '\\$&');
}

// baseName returns the last path segment of a file path (works for POSIX
// mounts), trimming a trailing slash first. Used to show a filename when no
// title metadata is available.
function baseName(p) {
  if (!p) return '';
  return String(p).replace(/\/+$/, '').split('/').pop();
}

// ---- Lane helpers ----
function laneLabel(name) {
  const meta = state.sources.find(s => s.name === name);
  if (meta && meta.label) return meta.label;
  if (SOURCE_COLORS[name] && SOURCE_COLORS[name].label) return SOURCE_COLORS[name].label;
  return name || 'Unknown';
}

// Ordered list of lane keys: SourceMeta order first, then any extra sources
// that showed up in results but weren't in the meta.
function laneOrder() {
  const names = state.sources.map(s => s.name);
  const set = new Set(names);
  for (const r of state.searchResults) {
    const k = laneKeyFor(r.source);
    if (!set.has(k)) { set.add(k); names.push(k); }
  }
  return names;
}

function laneResults(name) {
  return sortResults(state.searchResults.filter(r => laneKeyFor(r.source) === name));
}

function laneStatus(name) {
  return state.sourceStatus.get(name) || { status: 'done', count: 0, timeMs: null };
}

// Recomputes each lane's result count into sourceStatus.
function computeLaneCounts() {
  const counts = new Map();
  for (const r of state.searchResults) {
    const k = laneKeyFor(r.source);
    counts.set(k, (counts.get(k) || 0) + 1);
  }
  for (const [name, status] of state.sourceStatus.entries()) {
    status.count = counts.get(name) || 0;
  }
}

// ---- Render router ----
function renderSearchResults() {
  hideSearchSkeleton();
  computeLaneCounts();
  document.getElementById('search-no-results').classList.add('hidden');
  syncToolbarButtons();
  const container = document.getElementById('search-results');
  if (state.viewGroup === 'source') {
    renderSwimlanes(container);
  } else {
    renderFlat(container);
  }
}

function renderFlat(container) {
  const sorted = sortResults(state.searchResults);
  container.className = 'mb-6';
  if (state.viewLayout === 'list') {
    container.innerHTML = renderResultsTable(sorted);
  } else {
    container.innerHTML = `<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">${sorted.map((r, i) => renderBookCard(r, i)).join('')}</div>`;
  }
}

function renderSwimlanes(container) {
  container.className = 'space-y-3 mb-6';
  container.innerHTML = laneOrder().map(name => renderLane(name)).join('');
}

function renderLane(name) {
  const results = laneResults(name);
  const collapsed = state.collapsedLanes.has(name) || (laneStatus(name).status !== 'searching' && results.length === 0);
  // Persist the auto-collapse decision so the chevron/state stays consistent.
  if (collapsed) state.collapsedLanes.add(name);
  return `
    <div class="search-lane bg-slate-900/40 rounded-xl border border-slate-800 overflow-hidden" data-lane="${escapeHtml(name)}">
      ${renderLaneHeader(name, results.length, collapsed)}
      <div data-lane-body class="p-3 ${collapsed ? 'hidden' : ''}">${renderLaneBody(name, results)}</div>
    </div>
  `;
}

function renderLaneHeader(name, count, collapsed) {
  const status = laneStatus(name);
  const src = SOURCE_COLORS[name] || { bg: '#475569', text: 'white' };
  const time = (status.timeMs != null && status.timeMs > 0)
    ? `<span class="text-xs text-slate-500">${(status.timeMs / 1000).toFixed(1)}s</span>` : '';
  return `
    <button data-action="toggleLane" data-arg="${escapeHtml(name)}"
      class="w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-slate-800/50 transition-colors">
      <svg data-lane-chevron class="w-4 h-4 text-slate-500 transition-transform ${collapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
      <span class="px-2 py-0.5 rounded text-xs font-medium" style="background:${src.bg};color:${src.text || 'white'}">${escapeHtml(laneLabel(name))}</span>
      ${renderLaneBadge(status, count)}
      <span class="ml-auto flex items-center gap-2">${time}</span>
    </button>
  `;
}

function renderLaneBadge(status, count) {
  switch (status.status) {
    case 'searching':
      return `<span class="text-xs text-indigo-400 flex items-center gap-1"><svg class="w-3 h-3 spin" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="3"></circle><path class="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-2.5A6.5 6.5 0 0 0 12 5.5V3z"></path></svg>${t('lane_searching')}</span>`;
    case 'circuit':
      return `<span class="text-xs text-amber-400">⚠ ${t('lane_circuit')}</span>`;
    case 'error':
      return `<span class="text-xs text-rose-400" title="${escapeHtml(status.error || '')}">✗ ${t('lane_error')}</span>`;
    default:
      return `<span class="text-xs text-slate-400">${t('lane_done', {n: count})}</span>`;
  }
}

function renderLaneBody(name, results) {
  const status = laneStatus(name);
  if (results.length === 0) {
    if (status.status === 'searching') return laneSkeletonHtml();
    if (status.status === 'error') return `<p class="text-xs text-rose-400/80">${escapeHtml(status.error || t('lane_error'))}</p>`;
    return `<p class="text-xs text-slate-500">${t('lane_done', {n: 0})}</p>`;
  }
  if (state.viewLayout === 'list') return renderResultsTable(results);
  return `<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">${results.map((r, i) => renderBookCard(r, i)).join('')}</div>`;
}

function laneSkeletonHtml() {
  let html = '<div class="space-y-2">';
  for (let i = 0; i < 3; i++) {
    html += `<div class="flex items-center gap-3"><div class="skeleton h-10 w-8 rounded"></div><div class="flex-1 space-y-1.5"><div class="skeleton h-3 w-2/3 rounded"></div><div class="skeleton h-3 w-1/3 rounded"></div></div></div>`;
  }
  return html + '</div>';
}

// Targeted update of a single lane during streaming — preserves the scroll
// position and collapse state of every other lane.
function applyLaneFrame(source) {
  const name = laneKeyFor(source);
  const lane = document.querySelector(`#search-results .search-lane[data-lane="${cssEscape(name)}"]`);
  if (!lane) { renderSearchResults(); return; }
  computeLaneCounts();
  const results = laneResults(name);
  // Auto-collapse a lane once it settles with zero results (transparency: keep
  // it visible but folded). Mirrors renderLane so streaming matches the final render.
  if (laneStatus(name).status !== 'searching' && results.length === 0) {
    state.collapsedLanes.add(name);
  }
  const collapsed = state.collapsedLanes.has(name);
  const tmp = document.createElement('div');
  tmp.innerHTML = renderLaneHeader(name, results.length, collapsed);
  lane.querySelector('button[data-action="toggleLane"]').replaceWith(tmp.firstElementChild);
  const body = lane.querySelector('[data-lane-body]');
  body.classList.toggle('hidden', collapsed);
  body.innerHTML = renderLaneBody(name, results);
}

// Refreshes the count badges + response times on every lane (cross-source
// dedup can shift counts on lanes other than the one targeted this frame).
function refreshLaneCounts() {
  computeLaneCounts();
  for (const name of laneOrder()) {
    const lane = document.querySelector(`#search-results .search-lane[data-lane="${cssEscape(name)}"]`);
    if (!lane) { renderSearchResults(); return; }
    const collapsed = state.collapsedLanes.has(name);
    const tmp = document.createElement('div');
    tmp.innerHTML = renderLaneHeader(name, laneStatus(name).count, collapsed);
    lane.querySelector('button[data-action="toggleLane"]').replaceWith(tmp.firstElementChild);
  }
}

// Reflects the current view prefs + sort in the toolbar: active toggle
// highlighting, and showing sort buttons only in grid layout (list sorts by
// clicking column headers instead).
function syncToolbarButtons() {
  document.querySelectorAll('#search-toolbar [data-view-layout]').forEach(el => {
    const active = el.dataset.viewLayout === state.viewLayout;
    el.className = `seg-btn text-xs px-3 py-1 rounded-md transition-colors ${active ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-white'}`;
  });
  document.querySelectorAll('#search-toolbar [data-view-group]').forEach(el => {
    const active = el.dataset.viewGroup === state.viewGroup;
    el.className = `seg-btn text-xs px-3 py-1 rounded-md transition-colors ${active ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-white'}`;
  });
  const sortButtons = document.getElementById('search-sort-buttons');
  if (sortButtons) sortButtons.classList.toggle('hidden', state.viewLayout !== 'grid');
  document.querySelectorAll('#search-sort-buttons .sort-btn').forEach(el => {
    const active = el.dataset.sort === state.sortMode;
    el.className = `sort-btn text-xs px-3 py-1 rounded-md transition-colors ${active ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'}`;
  });
}

// ---- Table rendering ----
function renderResultsTable(results) {
  return `
    <div class="overflow-x-auto">
      <table class="w-full text-sm border-collapse">
        ${renderTableHead()}
        <tbody>${results.map((r, i) => renderBookRow(r, i)).join('')}</tbody>
      </table>
    </div>
  `;
}

function renderTableHead() {
  const cols = [
    { key: 'title', label: t('col_title') },
    { key: 'author', label: t('col_author') },
    { key: 'source', label: t('col_source') },
    { key: 'format', label: t('col_format') },
    { key: 'size', label: t('col_size') },
    { key: 'seeders', label: t('col_seeders') },
    { key: 'indexer', label: t('col_indexer') },
  ];
  const arrow = key => state.sortMode === key ? (state.sortDir === 'asc' ? ' ▲' : ' ▼') : '';
  const th = cols.map(c => `
    <th data-action="setSortMode" data-arg="${c.key}"
      class="text-left font-medium text-slate-400 px-2 py-1.5 cursor-pointer hover:text-white whitespace-nowrap select-none ${state.sortMode === c.key ? 'text-white' : ''}">
      ${escapeHtml(c.label)}${arrow(c.key)}
    </th>`).join('');
  return `<thead><tr class="border-b border-slate-800">
    <th class="px-2 py-1.5"></th>${th}
    <th class="px-2 py-1.5 text-right font-medium text-slate-400 whitespace-nowrap">${escapeHtml(t('col_action'))}</th>
  </tr></thead>`;
}

function renderBookRow(result, index) {
  const coverHtml = result.cover_url
    ? `<img src="${escapeHtml(result.cover_url)}" alt="" class="h-12 w-9 object-cover rounded" loading="lazy" data-ph-title="${escapeHtml(result.title || '')}" data-ph-idx="${index}" data-ph-size="sm">`
    : makeSmallPlaceholderHtml(result.title || '?', index);
  const seeders = result.seeders ? `<span class="text-emerald-400">${escapeHtml(String(result.seeders))}</span>` : '';
  const leechers = result.leechers ? `<span class="text-amber-400"> / ${escapeHtml(String(result.leechers))}</span>` : '';
  const swarm = (seeders || leechers) ? `${seeders}${leechers}` : '<span class="text-slate-600">—</span>';
  const sizeText = resultSizeText(result);
  return `
    <tr class="border-b border-slate-800/60 hover:bg-slate-800/30">
      <td class="px-2 py-1.5">${coverHtml}</td>
      <td class="px-2 py-1.5 max-w-xs"><span class="text-white line-clamp-2" title="${escapeHtml(result.title || '')}">${escapeHtml(result.title || 'Unknown')}</span></td>
      <td class="px-2 py-1.5 text-slate-400 max-w-[10rem]"><span class="line-clamp-1">${escapeHtml(result.author || '')}</span></td>
      <td class="px-2 py-1.5 text-slate-400 whitespace-nowrap">${escapeHtml(laneLabel(laneKeyFor(result.source)))}</td>
      <td class="px-2 py-1.5 text-slate-500 uppercase whitespace-nowrap">${escapeHtml(result.format || '')}</td>
      <td class="px-2 py-1.5 text-slate-400 whitespace-nowrap">${escapeHtml(sizeText)}</td>
      <td class="px-2 py-1.5 whitespace-nowrap">${swarm}</td>
      <td class="px-2 py-1.5 text-slate-500 whitespace-nowrap">${escapeHtml(result.indexer || '')}</td>
      <td class="px-2 py-1.5 text-right">${renderDownloadButton(result, true)}</td>
    </tr>
  `;
}

// ---- Download button (shared by cards + rows) ----
function downloadButtonMeta(result) {
  const downloadKey = getDownloadKey(result);
  const isDownloading = state.pendingDownloads.has(downloadKey);
  const isTrackedAnnaDownload = hasTrackedAnnaDownload(downloadKey);
  const isTrackedDirectDownload = hasTrackedDirectDownload(downloadKey);
  const trackedJob = getTrackedDownloadJob(downloadKey);
  const downloadOutcome = state.downloadOutcomes.get(downloadKey);
  const buttonState = (isDownloading || isTrackedAnnaDownload || isTrackedDirectDownload)
    ? 'loading' : (downloadOutcome ? downloadOutcome.status : 'idle');
  const text = {
    idle: t('download'), loading: t('loading'),
    success: t('download_added'), error: t('download_failed_state'),
  };
  const displayText = buttonState === 'loading' && trackedJob?.detail
    ? escapeHtml(trackedJob.detail) : (text[buttonState] || text.idle);
  return { buttonState, displayText };
}

function renderDownloadButton(result, compact = false) {
  const { buttonState, displayText } = downloadButtonMeta(result);
  const styles = {
    idle: 'bg-indigo-600 hover:bg-indigo-500 text-white',
    loading: 'bg-indigo-500/70 text-white cursor-not-allowed',
    success: 'bg-emerald-600 hover:bg-emerald-500 text-white cursor-default',
    error: 'bg-rose-600 hover:bg-rose-500 text-white cursor-pointer',
  };
  const icon = {
    idle: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>`,
    loading: `<svg class="w-4 h-4 spin" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2"></circle><path class="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-2.5A6.5 6.5 0 0 0 12 5.5V3z"></path></svg>`,
    success: `<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/></svg>`,
    error: `<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 8v4m0 4h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z"/></svg>`,
  };
  const disabled = (buttonState === 'idle' || buttonState === 'error') ? '' : 'disabled aria-busy="true"';
  const sizing = compact
    ? 'text-xs px-3 py-1 gap-1'
    : 'w-full text-sm py-1.5 gap-1.5';
  return `
    <button data-action="startDownload" data-key="${escapeHtml(result.__key || '')}" ${disabled}
      class="${sizing} ${styles[buttonState] || styles.idle} font-medium rounded-lg transition-colors inline-flex items-center justify-center disabled:opacity-100">
      ${icon[buttonState] || icon.idle}
      <span class="truncate">${displayText}</span>
    </button>`;
}

// Refreshes only the download buttons in place (no re-render) so downloads
// don't reset scroll or lane state.
function syncDownloadButtons() {
  const buttons = document.querySelectorAll('#search-results button[data-action="startDownload"]');
  buttons.forEach(btn => {
    const result = state.resultsByKey.get(btn.dataset.key);
    if (!result) return;
    const compact = btn.classList.contains('text-xs');
    const tmp = document.createElement('div');
    tmp.innerHTML = renderDownloadButton(result, compact);
    btn.replaceWith(tmp.firstElementChild);
  });
}

function renderBookCard(result, index) {
  const src = SOURCE_COLORS[result.source] || { bg: '#475569', text: 'white', label: result.source || 'Unknown' };
  const coverHtml = result.cover_url
    ? `<img src="${escapeHtml(result.cover_url)}" alt="" class="w-full h-48 object-cover" loading="lazy" data-ph-title="${escapeHtml(result.title || '')}" data-ph-idx="${index}">`
    : makePlaceholderHtml(result.title || '?', index);

  const seeders = result.seeders ? `<span class="text-emerald-400 text-xs font-medium">${t('n_seeds', {n: result.seeders})}</span>` : '';
  const leechers = result.leechers ? `<span class="text-amber-400 text-xs">${t('n_leech', {n: result.leechers})}</span>` : '';
  const sizeText = resultSizeText(result);
  const size = sizeText ? `<span class="text-slate-400 text-xs font-medium">${escapeHtml(sizeText)}</span>` : '';
  const format = result.format ? `<span class="text-slate-500 text-xs uppercase">${escapeHtml(result.format)}</span>` : '';
  const indexer = result.indexer ? `<span class="text-slate-600 text-xs">${escapeHtml(result.indexer)}</span>` : '';

  return `
    <div class="book-card bg-slate-900 rounded-xl overflow-hidden border border-slate-800 hover:border-slate-600 flex flex-col">
      <div class="relative">
        ${coverHtml}
        <span class="absolute top-2 left-2 px-2 py-0.5 rounded text-xs font-medium" style="background:${src.bg};color:${src.text}">${escapeHtml(src.label)}</span>
      </div>
      <div class="p-3 flex-1 flex flex-col">
        <h3 class="text-sm font-semibold text-white line-clamp-2 mb-1" title="${escapeHtml(result.title || '')}">${escapeHtml(result.title || 'Unknown')}</h3>
        <p class="text-xs text-slate-400 mb-2 line-clamp-1">${escapeHtml(result.author || '')}</p>
        <div class="flex items-center gap-2 flex-wrap mt-auto mb-2">
          ${seeders}${leechers}${size}${format}${indexer}
        </div>
        ${renderDownloadButton(result, false)}
      </div>
    </div>
  `;
}

function makeSmallPlaceholderHtml(title, index) {
  const gradient = COVER_GRADIENTS[index % COVER_GRADIENTS.length];
  const letter = (title || '?').charAt(0).toUpperCase();
  return `<div class="h-12 w-9 rounded bg-gradient-to-br ${gradient} flex items-center justify-center text-white text-xs font-bold">${escapeHtml(letter)}</div>`;
}

function makePlaceholderHtml(title, index) {
  const gradient = COVER_GRADIENTS[index % COVER_GRADIENTS.length];
  const letter = (title || '?').charAt(0).toUpperCase();
  return `<div class="w-full h-48 bg-gradient-to-br ${gradient} cover-placeholder">${escapeHtml(letter)}</div>`;
}

// Global function for img onerror fallback
window.makePlaceholder = function(title, index) {
  const gradient = COVER_GRADIENTS[index % COVER_GRADIENTS.length];
  const letter = (title || '?').charAt(0).toUpperCase();
  return `<div class="w-full h-48 bg-gradient-to-br ${gradient} cover-placeholder">${escapeHtml(letter)}</div>`;
};

function formatSize(size) {
  if (typeof size === 'string') return size;
  if (typeof size === 'number') {
    if (size > 1073741824) return (size / 1073741824).toFixed(1) + ' GB';
    if (size > 1048576) return (size / 1048576).toFixed(1) + ' MB';
    if (size > 1024) return (size / 1024).toFixed(1) + ' KB';
    return size + ' B';
  }
  return '';
}

function getDownloadKey(result) {
  return [
    result.source || '',
    result.download_url || '',
    result.url || '',
    result.abb_url || '',
    result.info_hash || '',
    result.magnet || '',
    result.md5 || '',
    result.source_id || '',
    result.title || '',
    result.author || '',
  ].join('|');
}

function hasTrackedAnnaDownload(downloadKey) {
  for (const tracked of state.trackedDownloadJobs.values()) {
    if (tracked.key === downloadKey && tracked.source === 'annas') return true;
  }
  return false;
}

function hasTrackedDirectDownload(downloadKey) {
  for (const tracked of state.trackedDownloadJobs.values()) {
    if (tracked.key === downloadKey && tracked.source !== 'annas') return true;
  }
  return false;
}

function getTrackedDownloadJob(downloadKey) {
  for (const [jobId, tracked] of state.trackedDownloadJobs.entries()) {
    if (tracked.key !== downloadKey) continue;
    const job = state.downloadJobs.find(candidate => String(candidate.job_id) === String(jobId));
    if (job) return job;
  }
  return null;
}

function setDownloadOutcome(downloadKey, status, persist = false) {
  const prevTimer = state.downloadOutcomeTimers.get(downloadKey);
  if (prevTimer) clearTimeout(prevTimer);

  state.downloadOutcomes.set(downloadKey, { status });
  syncDownloadButtons();

  if (persist) return;

  const timer = setTimeout(() => {
    state.downloadOutcomes.delete(downloadKey);
    state.downloadOutcomeTimers.delete(downloadKey);
    syncDownloadButtons();
  }, 2500);
  state.downloadOutcomeTimers.set(downloadKey, timer);
}

function isTerminalDownloadStatus(status) {
  return status === 'completed' || status === 'error' || status === 'dead_letter';
}

function isActiveDownloadStatus(status) {
  return status === 'queued' || status === 'searching' || status === 'downloading' || status === 'organizing' || status === 'importing' || status === 'retry_wait';
}

function trackDownloadJob(downloadKey, jobId, title, source, url = '') {
  if (!jobId) return;
  state.trackedDownloadJobs.set(String(jobId), { key: downloadKey, title, source, url });
  startDownloadPolling();
  refreshDownloads();
  syncDownloadButtons();
}

// ============================================================
// DOWNLOAD
// ============================================================
async function startDownload(result) {
  const downloadKey = getDownloadKey(result);
  if (state.pendingDownloads.has(downloadKey)) return;

  state.pendingDownloads.add(downloadKey);
  syncDownloadButtons();

  try {
    const body = {
      title: result.title,
      // Direct-download sources (Gutenberg, Standard Ebooks) carry their
      // link in epub_url — without this fallback their Download button 400s.
      download_url: result.download_url || result.url || result.epub_url || '',
      abb_url: result.abb_url || '',
      source: result.source,
      source_id: result.source_id || '',
      md5: result.md5 || '',
      author: result.author || '',
      info_hash: result.info_hash || '',
      magnet: result.magnet || '',
    };

    const data = await apiJson('/api/download', {
      method: 'POST',
      body: JSON.stringify(body),
    });

    if (result.source === 'annas' && data.job_id) {
      setDownloadOutcome(downloadKey, 'loading', true);
      trackAnnaDownload(data.job_id, downloadKey, result.title, result.download_url || result.url || '');
    } else if (data.success || data.job_id) {
      if (data.job_id) {
        setDownloadOutcome(downloadKey, 'loading', true);
        trackDownloadJob(downloadKey, data.job_id, result.title, result.source, result.download_url || result.url || '');
        showToast(t('download_started', {title: result.title}), 'info');
      } else {
        setDownloadOutcome(downloadKey, 'success');
        showToast(t('download_started', {title: result.title}), 'success');
      }
    } else {
      if (result.source === 'annas' && isAnnaNoMatchError(data.error || '')) {
        setDownloadOutcome(downloadKey, 'error', true);
        showAnnaNoMatchToast(result.title, result.download_url || result.url || '');
      } else {
        setDownloadOutcome(downloadKey, 'error');
        showToast(t('download_failed', {msg: data.error || t('unknown_error')}), 'error');
      }
    }
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      if (result.source === 'annas' && isAnnaNoMatchError(err.message || '')) {
        setDownloadOutcome(downloadKey, 'error', true);
        showAnnaNoMatchToast(result.title, result.download_url || result.url || '');
      } else {
        setDownloadOutcome(downloadKey, 'error');
        showToast(t('download_failed', {msg: err.message}), 'error');
      }
    }
  } finally {
    state.pendingDownloads.delete(downloadKey);
    syncDownloadButtons();
  }
}

function setDownloadsRefreshLoading(loading) {
  const button = document.getElementById('downloads-refresh-btn');
  const icon = document.getElementById('downloads-refresh-icon');
  if (!button || !icon) return;

  button.disabled = loading;
  if (loading) {
    button.setAttribute('aria-busy', 'true');
    icon.classList.add('spin');
  } else {
    button.removeAttribute('aria-busy');
    icon.classList.remove('spin');
  }
}

async function refreshDownloads(manual = false) {
  if (manual) setDownloadsRefreshLoading(true);

  try {
    const data = await apiJson('/api/downloads');
    state.downloadJobs = data.downloads || data.jobs || [];
    syncTrackedDownloadJobs(state.downloadJobs);
    renderDownloadList();
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      showToast(t('failed_load_downloads'), 'error');
    }
  } finally {
    if (manual) setDownloadsRefreshLoading(false);
  }
}

function renderDownloadList() {
  const jobs = state.downloadJobs || [];
  const container = document.getElementById('downloads-list');
  const emptyEl = document.getElementById('downloads-empty');

  // Update badge
  const activeCount = jobs.filter(j => isActiveDownloadStatus(j.status)).length;
  const badge = document.getElementById('dl-badge');
  if (activeCount > 0) {
    badge.textContent = activeCount;
    badge.classList.remove('hidden');
  } else {
    badge.classList.add('hidden');
  }

  if (jobs.length === 0) {
    container.innerHTML = '';
    emptyEl.classList.remove('hidden');
    return;
  }

  emptyEl.classList.add('hidden');
  container.innerHTML = jobs.map(renderDownloadJob).join('');
}

function renderDownloadJob(job) {
  const st = STATUS_STYLES[job.status] || STATUS_STYLES.queued;
  const progress = job.progress || 0;
  const showProgress = job.status === 'downloading' && progress > 0;
  const retryKey = String(job.job_id);
  const retryPending = state.pendingRetryDownloads.has(retryKey);

  let actions = '';
  if (job.status === 'error' || job.status === 'dead_letter') {
    actions = `
      <button
        data-action="retryDownload" data-job-id="${escapeHtml(job.job_id)}"
        ${retryPending ? 'disabled aria-busy="true"' : ''}
        class="px-2.5 py-1 text-xs bg-slate-700 hover:bg-slate-600 text-slate-300 rounded transition-colors flex items-center gap-1 disabled:opacity-100"
      >
        ${retryPending ? '<svg class="w-3.5 h-3.5 spin" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2"></circle><path class="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-2.5A6.5 6.5 0 0 0 12 5.5V3z"></path></svg>' : ''}
        <span>${retryPending ? t('loading') : t('retry')}</span>
      </button>`;
  }

  return `
    <div class="bg-slate-900 border ${st.border} rounded-xl p-4 flex flex-col sm:flex-row sm:items-center gap-3">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <span class="px-2 py-0.5 rounded text-xs font-medium ${st.bg} ${st.text}">${t('status_' + job.status)}</span>
          ${job.source ? `<span class="text-xs text-slate-500">${escapeHtml(job.source)}</span>` : ''}
        </div>
        <h4 class="text-sm font-medium text-white truncate" title="${escapeHtml(job.title || '')}">${escapeHtml(job.title || 'Unknown')}</h4>
        ${job.detail ? `<p class="text-xs text-slate-400 mt-1 truncate" title="${escapeHtml(job.detail)}">${escapeHtml(job.detail)}</p>` : ''}
        ${job.max_retries > 0 && job.retry_count > 0 ? `<p class="text-xs text-amber-400 mt-1">${escapeHtml(`Attempt ${Math.min(job.retry_count + 1, job.max_retries + 1)}/${job.max_retries + 1}`)}</p>` : ''}
        ${job.error ? `<p class="text-xs text-red-400 mt-1 truncate">${escapeHtml(job.error)}</p>` : ''}
        ${showProgress ? `
          <div class="mt-2 w-full bg-slate-800 rounded-full h-1.5">
            <div class="progress-bar bg-indigo-500 h-1.5 rounded-full" style="width:${progress}%"></div>
          </div>
          <span class="text-xs text-slate-500 mt-1">${progress.toFixed(1)}%</span>
        ` : ''}
      </div>
      <div class="flex items-center gap-2 flex-shrink-0">
        ${job.size ? `<span class="text-xs text-slate-500">${escapeHtml(formatSize(job.size))}</span>` : ''}
        ${actions}
      </div>
    </div>
  `;
}

function syncTrackedDownloadJobs(jobs) {
  const jobsById = new Map((jobs || []).map(job => [String(job.job_id), job]));
  let hasPendingTrackedJob = false;

  for (const [jobId, tracked] of [...state.trackedDownloadJobs.entries()]) {
    const job = jobsById.get(jobId);
    if (!job) {
      hasPendingTrackedJob = true;
      continue;
    }

    if (job.status === 'completed') {
      state.trackedDownloadJobs.delete(jobId);
      setDownloadOutcome(tracked.key, 'success');
      showToast(t('download_complete', {title: tracked.title || job.title || t('unknown_title')}), 'success');
      continue;
    }

    if (TERMINAL_DOWNLOAD_STATUSES.has(job.status)) {
      state.trackedDownloadJobs.delete(jobId);
      if (isAnnaNoMatchError(job.error || '')) {
        setDownloadOutcome(tracked.key, 'error', true);
        showAnnaNoMatchToast(tracked.title || job.title || 'Unknown', tracked.url || '');
      } else {
        setDownloadOutcome(tracked.key, 'error');
        showToast(t('download_failed', {msg: job.error || t('unknown_error')}), 'error');
      }
      continue;
    }

    hasPendingTrackedJob = true;
    Object.assign(tracked, {
      status: job.status,
      detail: job.detail || '',
      retryCount: job.retry_count || 0,
      maxRetries: job.max_retries || 0,
    });
    setDownloadOutcome(tracked.key, 'loading', true);
  }

  const hasActiveJob = (jobs || []).some(job => isActiveDownloadStatus(job.status));
  if (!hasPendingTrackedJob && state.trackedDownloadJobs.size === 0 && !hasActiveJob) {
    stopDownloadPolling();
  }
}

function trackAnnaDownload(jobId, downloadKey, title, url) {
  for (const [trackedJobId, tracked] of [...state.trackedDownloadJobs.entries()]) {
    if (tracked.key === downloadKey) {
      state.trackedDownloadJobs.delete(trackedJobId);
    }
  }
  state.trackedDownloadJobs.set(String(jobId), { key: downloadKey, title, url: url || '', source: 'annas' });
  startDownloadPolling();
  refreshDownloads();
}

async function retryDownload(jobId) {
  const key = String(jobId);
  if (state.pendingRetryDownloads.has(key)) return;

  state.pendingRetryDownloads.add(key);
  renderDownloadList();

  try {
    await apiJson(`/api/downloads/jobs/${jobId}/retry`, { method: 'POST' });
    showToast(t('retrying_download'), 'info');
    await refreshDownloads();
  } catch (err) {
    if (err.message !== 'Unauthorized') showToast(t('retry_failed'), 'error');
  } finally {
    state.pendingRetryDownloads.delete(key);
    renderDownloadList();
  }
}

async function clearCompleted() {
  const button = document.getElementById('downloads-clear-btn');
  const icon = document.getElementById('downloads-clear-icon');
  if (button && icon) {
    button.disabled = true;
    button.setAttribute('aria-busy', 'true');
    icon.classList.remove('hidden');
  }

  try {
    await apiJson('/api/downloads/clear', { method: 'POST' });
    showToast(t('cleared_completed'), 'success');
    await refreshDownloads();
  } catch (err) {
    if (err.message !== 'Unauthorized') showToast(t('failed_clear'), 'error');
  } finally {
    if (button && icon) {
      button.disabled = false;
      button.removeAttribute('aria-busy');
      icon.classList.add('hidden');
    }
  }
}

function startDownloadPolling() {
  stopDownloadPolling();
  state.downloadPollTimer = setInterval(refreshDownloads, 5000);
}

function stopDownloadPolling() {
  if (state.downloadPollTimer) {
    clearInterval(state.downloadPollTimer);
    state.downloadPollTimer = null;
  }
}

// ============================================================
// LIBRARY
// ============================================================
let librarySearchTimeout = null;

document.getElementById('library-search').addEventListener('input', (e) => {
  clearTimeout(librarySearchTimeout);
  state.libraryPage = 1;
  librarySearchTimeout = setTimeout(() => loadLibrary(), 400);
});

async function loadLibrary() {
  const tab = state.libraryTab;
  const q = document.getElementById('library-search').value.trim();
  const page = state.libraryPage;

  const endpoints = {
    ebooks: `/api/library?page=${page}${q ? '&q=' + encodeURIComponent(q) : ''}`,
    audiobooks: `/api/library/audiobooks?page=${page}${q ? '&q=' + encodeURIComponent(q) : ''}`,
    manga: `/api/library/manga?page=${page}${q ? '&q=' + encodeURIComponent(q) : ''}`,
  };

  const container = document.getElementById('library-results');
  const emptyEl = document.getElementById('library-empty');
  const paginationEl = document.getElementById('library-pagination');

  try {
    const data = await apiJson(endpoints[tab]);
    const items = data.items || [];
    state.libraryPages = data.pages || 1;

    if (items.length === 0) {
      container.innerHTML = '';
      emptyEl.classList.remove('hidden');
      paginationEl.classList.add('hidden');
      return;
    }

    emptyEl.classList.add('hidden');

    // Group items by series and render with section headers
    const renderFn = tab === 'ebooks' ? renderLibraryEbook : tab === 'audiobooks' ? renderLibraryAudiobook : renderLibraryManga;
    container.innerHTML = renderGroupedBySeries(items, renderFn);

    // Pagination
    if (state.libraryPages > 1) {
      paginationEl.classList.remove('hidden');
      renderPagination(paginationEl, page, state.libraryPages);
    } else {
      paginationEl.classList.add('hidden');
    }
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      container.innerHTML = `<div class="col-span-full text-center py-10 text-slate-500">${t('failed_load_library')}</div>`;
    }
  }
}

function seriesBaseName(series) {
  // Strip "#N", "Book N", "(N)", "Vol N", "Volume N" suffixes to get the base series name
  if (!series) return '';
  return series
    .replace(/\s*#\s*[\d.]+.*$/, '')
    .replace(/\s*Book\s+[\d.]+.*$/i, '')
    .replace(/\s*Vol(ume)?\s*\.?\s*[\d.]+.*$/i, '')
    .replace(/\s*\([\d.]+\).*$/, '')
    .replace(/\s*,\s*$/, '')
    .trim();
}

function renderGroupedBySeries(items, renderFn) {
  const groups = new Map(); // base series name → items
  const standalone = [];
  let idx = 0;

  for (const item of items) {
    const raw = item.series || '';
    const base = seriesBaseName(raw);
    if (base) {
      if (!groups.has(base)) groups.set(base, []);
      groups.get(base).push(item);
    } else {
      standalone.push(item);
    }
  }

  let html = '';
  for (const [series, groupItems] of groups) {
    if (groupItems.length > 1) {
      // Series header
      html += `<div class="col-span-full mt-4 mb-2 first:mt-0">
        <div class="flex items-center gap-3">
          <h3 class="text-sm font-semibold text-indigo-400">${escapeHtml(series)}</h3>
          <span class="text-xs text-slate-600">${t('n_items', {n: groupItems.length})}</span>
          <div class="flex-1 border-t border-slate-800"></div>
        </div>
      </div>`;
    }
    for (const item of groupItems) {
      html += renderFn(item, idx++);
    }
    if (groupItems.length > 1) {
      // Spacer after series group
      html += `<div class="col-span-full h-2"></div>`;
    }
  }

  // Standalone items (no series)
  if (standalone.length > 0 && groups.size > 0) {
    html += `<div class="col-span-full mt-4 mb-2">
      <div class="flex items-center gap-3">
        <h3 class="text-sm font-semibold text-slate-500">${t('other')}</h3>
        <div class="flex-1 border-t border-slate-800"></div>
      </div>
    </div>`;
  }
  for (const item of standalone) {
    html += renderFn(item, idx++);
  }

  return html;
}

function renderLibraryEbook(item, index) {
  const coverHtml = item.cover_url
    ? `<img src="${escapeHtml(item.cover_url)}" alt="" class="w-full h-48 object-cover" loading="lazy" data-ph-title="${escapeHtml(item.title || '')}" data-ph-idx="${index}">`
    : makePlaceholderHtml(item.title || '?', index);

  const format = item.format || (item.file_path || '').split('.').pop() || '';

  const seriesBadge = item.series ? `<p class="text-xs text-indigo-400 line-clamp-1 mb-1">${escapeHtml(item.series)}</p>` : '';

  return `
    <div class="book-card bg-slate-900 rounded-xl overflow-hidden border border-slate-800 hover:border-slate-600 relative group">
      ${coverHtml}
      <button data-action="deleteLibraryItem" data-id="${escapeHtml(item.id || '')}" data-type="book" data-title="${escapeHtml(item.title || '')}"
        class="absolute top-2 right-2 w-7 h-7 bg-red-600/80 hover:bg-red-500 text-white rounded-full text-xs opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center" title="Remove from library">&#x2715;</button>
      <div class="p-3">
        <h3 class="text-sm font-semibold text-white line-clamp-2 mb-1">${escapeHtml(item.title || 'Unknown')}</h3>
        <p class="text-xs text-slate-400 line-clamp-1 mb-1">${escapeHtml(item.author || '')}</p>
        ${seriesBadge}
        <div class="flex items-center gap-2 text-xs text-slate-500">
          ${format ? `<span class="uppercase">${escapeHtml(format)}</span>` : ''}
          ${item.size ? `<span>${escapeHtml(formatSize(item.size))}</span>` : ''}
        </div>
      </div>
    </div>
  `;
}

function renderLibraryAudiobook(item, index) {
  const coverHtml = item.cover_url
    ? `<img src="${escapeHtml(item.cover_url)}" alt="" class="w-full h-48 object-cover" loading="lazy" data-ph-title="${escapeHtml(item.title || '')}" data-ph-idx="${index}">`
    : makePlaceholderHtml(item.title || '?', index);

  const duration = item.duration_hours ? `${item.duration_hours.toFixed(1)}h` : '';

  return `
    <div class="book-card bg-slate-900 rounded-xl overflow-hidden border border-slate-800 hover:border-slate-600 relative group">
      ${coverHtml}
      <button data-action="deleteLibraryItem" data-id="${escapeHtml(item.id || '')}" data-type="audiobook" data-title="${escapeHtml(item.title || '')}"
        class="absolute top-2 right-2 w-7 h-7 bg-red-600/80 hover:bg-red-500 text-white rounded-full text-xs opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center" title="Remove from library">&#x2715;</button>
      <div class="p-3">
        <h3 class="text-sm font-semibold text-white line-clamp-2 mb-1">${escapeHtml(item.title || 'Unknown')}</h3>
        <p class="text-xs text-slate-400 line-clamp-1 mb-1">${escapeHtml(item.author || '')}</p>
        ${item.series ? `<p class="text-xs text-indigo-400 line-clamp-1 mb-1">${escapeHtml(item.series)}</p>` : ''}
        <div class="flex items-center gap-2 text-xs text-slate-500 mb-2">
          ${duration ? `<span>${duration}</span>` : ''}
          ${item.num_files ? `<span>${t('n_files', {n: item.num_files})}</span>` : ''}
        </div>
        ${item.abs_url ? `<a href="${escapeHtml(item.abs_url)}" target="_blank" class="text-xs text-indigo-400 hover:text-indigo-300 transition-colors">${t('open_in_abs')}</a>` : ''}
      </div>
    </div>
  `;
}

function renderLibraryManga(item, index) {
  // Kavita rows carry `name`; local-DB fallback rows carry `title` (the torrent
  // name), or at worst just a file_path — derive a display name from whichever
  // is present so entries don't all show "Unknown".
  const displayName = item.name || item.title || baseName(item.file_path) || '';

  const coverHtml = item.cover_url
    ? `<img src="${escapeHtml(item.cover_url)}" alt="" class="w-full h-48 object-cover" loading="lazy" data-ph-title="${escapeHtml(displayName)}" data-ph-idx="${index}">`
    : makePlaceholderHtml(displayName || '?', index);

  const selected = state.selectedSeries.has(item.id);
  const selectClasses = state.selectMode ? 'cursor-pointer' : '';
  const ring = selected ? 'border-indigo-500 ring-2 ring-indigo-500/50' : 'border-slate-800 hover:border-slate-600';

  // In select mode the whole card toggles selection; a checkbox badge shows state.
  const selectOverlay = state.selectMode
    ? `<div class="absolute top-2 left-2 w-6 h-6 rounded-md flex items-center justify-center text-xs font-bold ${selected ? 'bg-indigo-600 text-white' : 'bg-slate-900/80 border border-slate-500 text-transparent'}">&#x2713;</div>`
    : '';
  const cardAttrs = state.selectMode
    ? `data-action="toggleSeriesSelect" data-id="${escapeHtml(String(item.id))}" data-name="${escapeHtml(displayName)}"`
    : '';

  return `
    <div class="book-card bg-slate-900 rounded-xl overflow-hidden border relative ${ring} ${selectClasses}" ${cardAttrs}>
      ${selectOverlay}
      ${coverHtml}
      <div class="p-3">
        <h3 class="text-sm font-semibold text-white line-clamp-2 mb-1">${escapeHtml(displayName || 'Unknown')}</h3>
        <div class="flex items-center gap-2 text-xs text-slate-500 mb-2">
          ${item.pages ? `<span>${t('n_pages', {n: item.pages})}</span>` : ''}
          ${item.library ? `<span>${escapeHtml(item.library)}</span>` : ''}
        </div>
        ${!state.selectMode && item.kavita_url ? `<a href="${escapeHtml(item.kavita_url)}" target="_blank" class="text-xs text-indigo-400 hover:text-indigo-300 transition-colors">${t('open_in_kavita')}</a>` : ''}
      </div>
    </div>
  `;
}

// ============================================================
// BULK METADATA EDITING (Kavita)
// ============================================================

function toggleSelectMode() {
  state.selectMode = !state.selectMode;
  const toggle = document.getElementById('library-select-toggle');
  toggle.classList.toggle('bg-indigo-600', state.selectMode);
  toggle.classList.toggle('text-white', state.selectMode);
  if (!state.selectMode) state.selectedSeries.clear();
  updateSelectionBar();
  loadLibrary();
}

function exitSelectMode() {
  if (!state.selectMode && state.selectedSeries.size === 0) return;
  state.selectMode = false;
  state.selectedSeries.clear();
  const toggle = document.getElementById('library-select-toggle');
  toggle.classList.remove('bg-indigo-600', 'text-white');
  updateSelectionBar();
}

function toggleSeriesSelect(id, name) {
  // Kavita series ids are integers; keep the map keyed numerically.
  const key = Number(id);
  if (state.selectedSeries.has(key)) {
    state.selectedSeries.delete(key);
  } else {
    state.selectedSeries.set(key, name);
  }
  updateSelectionBar();
  loadLibrary();
}

function clearSelection() {
  state.selectedSeries.clear();
  updateSelectionBar();
  loadLibrary();
}

function updateSelectionBar() {
  const bar = document.getElementById('library-selection-bar');
  const count = state.selectedSeries.size;
  document.getElementById('selection-count').textContent = count;
  if (state.selectMode && count > 0) {
    bar.classList.remove('hidden');
    bar.classList.add('flex');
  } else {
    bar.classList.add('hidden');
    bar.classList.remove('flex');
  }
}

async function openMetadataModal() {
  if (state.selectedSeries.size === 0) return;
  const modal = document.getElementById('metadata-modal');
  document.getElementById('metadata-modal-count').textContent = state.selectedSeries.size;
  document.getElementById('metadata-modal-error').classList.add('hidden');

  // Reset all rows.
  document.querySelectorAll('#metadata-fields .metadata-row').forEach(row => {
    const cb = row.querySelector('.meta-enable');
    const val = row.querySelector('.meta-value');
    cb.checked = false;
    val.disabled = true;
    val.value = '';
  });
  document.querySelector('input[name="meta-list-mode"][value="set"]').checked = true;

  // Renaming is opt-in per edit: never carry it over from a previous open.
  document.getElementById('metadata-rename-enable').checked = false;
  document.getElementById('metadata-rename-options').classList.add('hidden');
  document.getElementById('metadata-rename-preview').innerHTML = '';
  document.querySelector('input[name="meta-rename-mode"][value="folder"]').checked = true;

  modal.classList.remove('hidden');
  modal.classList.add('flex');

  // Online lookup is always available. Every editable field here is
  // series-level (author, genres, tags…), so applying one AniList result to
  // all selected series is valid — a multi-selection is the user asserting
  // these belong to the same series. The query prefills from the first
  // selected series' name; the user can edit it before searching.
  const online = document.getElementById('metadata-online');
  online.classList.remove('hidden');
  document.getElementById('metadata-online-results').innerHTML = '';
  document.getElementById('metadata-online-status').classList.add('hidden');
  const [, firstName] = [...state.selectedSeries.entries()][0];
  document.getElementById('metadata-online-query').value = firstName || '';
  if (state.selectedSeries.size === 1) prefillMetadataField('series', firstName || '');

  // Pre-fill existing values from Kavita only for a single-series edit — these
  // are read per-series and can't be meaningfully merged across a selection.
  if (state.selectedSeries.size === 1) {
    const [id] = state.selectedSeries.keys();
    try {
      const md = await apiJson(`/api/library/manga/series/${id}/metadata`);
      prefillMetadataField('authors', (md.authors || []).join(', '));
      prefillMetadataField('genres', (md.genres || []).join(', '));
      prefillMetadataField('tags', (md.tags || []).join(', '));
      prefillMetadataField('publisher', (md.publishers || [])[0] || '');
      prefillMetadataField('language', md.language || '');
      prefillMetadataField('summary', md.summary || '');
    } catch (err) {
      // Non-fatal: the form just stays blank if prefill fails.
    }
  }
}

function prefillMetadataField(field, value) {
  if (!value) return;
  const row = document.querySelector(`#metadata-fields .metadata-row[data-field="${field}"]`);
  if (row) row.querySelector('.meta-value').value = value;
}

// Holds the last online-search candidates so a click can prefill from the full
// object (rather than re-encoding it into the DOM).
let metadataCandidates = [];

async function searchOnlineMetadata() {
  const q = document.getElementById('metadata-online-query').value.trim();
  const type = document.getElementById('metadata-online-source').value;
  const statusEl = document.getElementById('metadata-online-status');
  const resultsEl = document.getElementById('metadata-online-results');
  if (!q) return;

  statusEl.textContent = t('searching') || 'Searching…';
  statusEl.classList.remove('hidden');
  resultsEl.innerHTML = '';
  metadataCandidates = [];

  try {
    const data = await apiJson(`/api/metadata/search?type=${encodeURIComponent(type)}&q=${encodeURIComponent(q)}`);
    metadataCandidates = data.candidates || [];
    if (metadataCandidates.length === 0) {
      statusEl.textContent = t('no_matches') || 'No matches found.';
      return;
    }
    statusEl.classList.add('hidden');
    resultsEl.innerHTML = metadataCandidates.map((c, i) => renderMetadataCandidate(c, i)).join('');
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      statusEl.textContent = err.message || t('search_failed') || 'Search failed.';
    }
  }
}

function renderMetadataCandidate(c, index) {
  const cover = c.cover_url
    ? `<img src="${escapeHtml(c.cover_url)}" alt="" class="h-14 w-10 object-cover rounded flex-shrink-0" loading="lazy">`
    : '<div class="h-14 w-10 rounded bg-slate-700 flex-shrink-0"></div>';
  const meta = [c.year, c.authors && c.authors.join(', ')].filter(Boolean).join(' · ');
  const link = c.site_url
    ? `<a href="${escapeHtml(c.site_url)}" target="_blank" rel="noopener noreferrer" class="text-indigo-400 hover:text-indigo-300 text-xs">↗</a>`
    : '';
  return `
    <div class="metadata-candidate flex items-start gap-3 p-2 rounded-lg border border-slate-700 hover:border-indigo-500 cursor-pointer transition-colors" data-action="applyMetadataCandidate" data-idx="${index}">
      ${cover}
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-1.5"><span class="text-sm font-medium text-white line-clamp-1">${escapeHtml(c.title || 'Unknown')}</span>${link}</div>
        <div class="text-xs text-slate-400 line-clamp-1">${escapeHtml(meta)}</div>
        <div class="text-xs text-slate-500 line-clamp-2 mt-0.5">${escapeHtml(c.summary || '')}</div>
      </div>
    </div>`;
}

// applyMetadataCandidate prefills the editable rows from a chosen candidate and
// enables each field it populated so the user can review before saving.
function applyMetadataCandidate(index) {
  const c = metadataCandidates[index];
  if (!c) return;
  const apply = (field, value) => {
    if (!value) return;
    const row = document.querySelector(`#metadata-fields .metadata-row[data-field="${field}"]`);
    if (!row) return;
    row.querySelector('.meta-value').value = value;
    const cb = row.querySelector('.meta-enable');
    cb.checked = true;
    row.querySelector('.meta-value').disabled = false;
  };
  apply('series', c.title || '');
  apply('authors', (c.authors || []).join(', '));
  apply('genres', (c.genres || []).join(', '));
  apply('tags', (c.tags || []).join(', '));
  apply('publisher', c.publisher || '');
  apply('language', c.language || '');
  apply('summary', c.summary || '');

  // Highlight the chosen candidate.
  document.querySelectorAll('.metadata-candidate').forEach(el => {
    el.classList.toggle('border-indigo-500', +el.dataset.idx === index);
    el.classList.toggle('bg-indigo-500/10', +el.dataset.idx === index);
  });
}

// --- Rename preview -------------------------------------------------------
//
// The preview calls the same planner the save path uses, so what is shown is
// what will run. It is recomputed whenever the inputs that feed it change.

let renamePreviewSeq = 0;

function selectedRenameMode() {
  if (!document.getElementById('metadata-rename-enable').checked) return 'none';
  const el = document.querySelector('input[name="meta-rename-mode"]:checked');
  return el ? el.value : 'folder';
}

function seriesFieldValue() {
  const row = document.querySelector('#metadata-fields .metadata-row[data-field="series"]');
  if (!row || !row.querySelector('.meta-enable').checked) return '';
  return row.querySelector('.meta-value').value.trim();
}

async function refreshRenamePreview() {
  const box = document.getElementById('metadata-rename-preview');
  const mode = selectedRenameMode();
  if (mode === 'none') {
    box.innerHTML = '';
    return;
  }

  const name = seriesFieldValue();
  if (!name) {
    box.innerHTML = `<span class="text-amber-400">${escapeHtml(t('rename_needs_series'))}</span>`;
    return;
  }

  // Late responses from a superseded request must not overwrite a newer preview.
  const seq = ++renamePreviewSeq;
  box.innerHTML = escapeHtml(t('rename_previewing'));

  try {
    // api() rather than apiJson() so a 4xx body's {error} reaches the user.
    const resp = await api('/api/library/metadata/rename-preview', {
      method: 'POST',
      body: JSON.stringify({
        series_ids: [...state.selectedSeries.keys()],
        series_name: name,
        mode,
      }),
    });
    const data = await resp.json().catch(() => ({}));
    if (seq !== renamePreviewSeq) return;
    if (!resp.ok) throw new Error(data.error || `API error: ${resp.status}`);
    box.innerHTML = renderRenamePreview(data.plans || []);
  } catch (err) {
    if (seq !== renamePreviewSeq) return;
    if (err.message !== 'Unauthorized') {
      box.innerHTML = `<span class="text-red-400">${escapeHtml(err.message || t('rename_preview_failed'))}</span>`;
    }
  }
}

function renderRenamePreview(plans) {
  const parts = [];
  let changes = 0;

  for (const p of plans) {
    const lines = [];
    if (p.folder_from && p.folder_to) {
      changes++;
      lines.push(renameLine(baseName(p.folder_from) + '/', baseName(p.folder_to) + '/'));
    }
    for (const f of p.files || []) {
      changes++;
      lines.push(renameLine(baseName(f.from), baseName(f.to)));
    }
    for (const skip of p.skipped || []) {
      lines.push(`<div class="text-slate-500">${escapeHtml(skip)}</div>`);
    }
    for (const warn of p.warnings || []) {
      lines.push(`<div class="text-amber-400">${escapeHtml(warn)}</div>`);
    }
    for (const e of p.errors || []) {
      lines.push(`<div class="text-red-400">${escapeHtml(e)}</div>`);
    }
    if (lines.length === 0) continue;
    if (plans.length > 1) {
      lines.unshift(`<div class="text-slate-300 mt-1">${escapeHtml(p.series_name || `#${p.series_id}`)}</div>`);
    }
    parts.push(lines.join(''));
  }

  if (parts.length === 0) return escapeHtml(t('rename_no_changes'));
  if (changes === 0) return parts.join('');
  return parts.join('');
}

function renameLine(from, to) {
  return `<div class="flex items-start gap-1 flex-wrap">
    <span class="text-slate-500 line-through">${escapeHtml(from)}</span>
    <span class="text-slate-600">→</span>
    <span class="text-emerald-400">${escapeHtml(to)}</span>
  </div>`;
}

function closeMetadataModal() {
  const modal = document.getElementById('metadata-modal');
  modal.classList.add('hidden');
  modal.classList.remove('flex');
}

async function saveMetadata() {
  const ids = [...state.selectedSeries.keys()];
  if (ids.length === 0) return;

  const listFields = new Set(['authors', 'genres', 'tags']);
  const fields = {};
  let anyEnabled = false;
  document.querySelectorAll('#metadata-fields .metadata-row').forEach(row => {
    if (!row.querySelector('.meta-enable').checked) return;
    anyEnabled = true;
    const key = row.dataset.field;
    const raw = row.querySelector('.meta-value').value;
    if (listFields.has(key)) {
      fields[key] = raw.split(',').map(s => s.trim()).filter(Boolean);
    } else {
      fields[key] = raw;
    }
  });

  const errEl = document.getElementById('metadata-modal-error');
  if (!anyEnabled) {
    errEl.textContent = t('metadata_no_fields') || 'Enable at least one field to write.';
    errEl.classList.remove('hidden');
    return;
  }

  const mode = document.querySelector('input[name="meta-list-mode"]:checked').value;
  const btn = document.getElementById('metadata-save-btn');
  btn.disabled = true;
  btn.textContent = t('writing') || 'Writing…';

  try {
    // Use api() directly so a 4xx body's {error} message reaches the user
    // instead of apiJson's generic "API error: 400".
    const resp = await api('/api/library/metadata/bulk', {
      method: 'POST',
      body: JSON.stringify({ series_ids: ids, mode, fields, rename: selectedRenameMode() }),
    });
    const res = await resp.json().catch(() => ({}));
    if (!resp.ok) {
      throw new Error(res.error || `API error: ${resp.status}`);
    }
    const fileErrors = (res.results || []).reduce((n, r) => n + (r.errors ? r.errors.length : 0), 0);
    if (fileErrors > 0) {
      showToast(
        `Wrote ${res.files_written} file(s) across ${res.series_count} series, ${fileErrors} issue(s)`,
        'warning'
      );
    } else {
      const renamed = res.paths_renamed ? `, renamed ${res.paths_renamed} path(s)` : '';
      showToast(`Wrote metadata to ${res.files_written} file(s) across ${res.series_count} series${renamed}`, 'success');
    }
    closeMetadataModal();
    exitSelectMode();
    loadLibrary();
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      errEl.textContent = err.message || 'Failed to write metadata';
      errEl.classList.remove('hidden');
    }
  } finally {
    btn.disabled = false;
    btn.textContent = t('save_write_files') || 'Save & write files';
  }
}

async function deleteLibraryItem(id, type, title) {
  if (!id) return;
  if (!confirm(`Remove "${title}" from library?`)) return;
  try {
    await apiJson(`/api/library/${type}/${id}`, { method: 'DELETE' });
    showToast(`Removed "${title}"`, 'success');
    loadLibrary();
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      showToast(`Failed to remove: ${err.message}`, 'error');
    }
  }
}

function renderPagination(container, currentPage, totalPages) {
  let html = '';

  // Previous
  html += `<button data-action="goLibraryPage" data-page="${currentPage - 1}" class="px-3 py-1.5 text-sm rounded-lg ${currentPage <= 1 ? 'bg-slate-800 text-slate-600 cursor-not-allowed' : 'bg-slate-800 hover:bg-slate-700 text-slate-300'}" ${currentPage <= 1 ? 'disabled' : ''}>${t('prev')}</button>`;

  // Page numbers
  const maxVisible = 7;
  let start = Math.max(1, currentPage - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages, start + maxVisible - 1);
  if (end - start < maxVisible - 1) start = Math.max(1, end - maxVisible + 1);

  if (start > 1) {
    html += `<button data-action="goLibraryPage" data-page="1" class="px-3 py-1.5 text-sm rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300">1</button>`;
    if (start > 2) html += `<span class="text-slate-600 px-1">...</span>`;
  }

  for (let i = start; i <= end; i++) {
    html += `<button data-action="goLibraryPage" data-page="${i}" class="px-3 py-1.5 text-sm rounded-lg ${i === currentPage ? 'bg-indigo-600 text-white' : 'bg-slate-800 hover:bg-slate-700 text-slate-300'}">${i}</button>`;
  }

  if (end < totalPages) {
    if (end < totalPages - 1) html += `<span class="text-slate-600 px-1">...</span>`;
    html += `<button data-action="goLibraryPage" data-page="${totalPages}" class="px-3 py-1.5 text-sm rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300">${totalPages}</button>`;
  }

  // Next
  html += `<button data-action="goLibraryPage" data-page="${currentPage + 1}" class="px-3 py-1.5 text-sm rounded-lg ${currentPage >= totalPages ? 'bg-slate-800 text-slate-600 cursor-not-allowed' : 'bg-slate-800 hover:bg-slate-700 text-slate-300'}" ${currentPage >= totalPages ? 'disabled' : ''}>${t('next')}</button>`;

  container.innerHTML = html;
}

function goLibraryPage(page) {
  if (page < 1 || page > state.libraryPages) return;
  state.libraryPage = page;
  loadLibrary();
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

// ============================================================
// WISHLIST
// ============================================================
async function loadWishlist() {
  try {
    const data = await apiJson('/api/wishlist');
    const items = data.items || [];
    const container = document.getElementById('wishlist-list');
    const emptyEl = document.getElementById('wishlist-empty');

    if (items.length === 0) {
      container.innerHTML = '';
      emptyEl.classList.remove('hidden');
      return;
    }

    emptyEl.classList.add('hidden');
    container.innerHTML = items.map(renderWishlistItem).join('');
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      showToast(t('failed_load_wishlist'), 'error');
    }
  }
}

function renderWishlistItem(item) {
  const typeColors = { ebook: 'bg-indigo-500/20 text-indigo-400', audiobook: 'bg-purple-500/20 text-purple-400', manga: 'bg-pink-500/20 text-pink-400' };
  const tc = typeColors[item.media_type] || typeColors.ebook;
  const date = item.created_at ? new Date(item.created_at).toLocaleDateString() : '';

  return `
    <div class="bg-slate-900 border border-slate-800 rounded-xl px-4 py-3 flex items-center gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-0.5">
          <span class="px-2 py-0.5 rounded text-xs font-medium ${tc}">${escapeHtml(item.media_type || 'ebook')}</span>
          ${date ? `<span class="text-xs text-slate-600">${date}</span>` : ''}
        </div>
        <h4 class="text-sm font-medium text-white truncate">${escapeHtml(item.title || '')}</h4>
        ${item.author ? `<p class="text-xs text-slate-400">${escapeHtml(item.author)}</p>` : ''}
      </div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <button data-action="searchWishlistItem" data-title="${escapeHtml(item.title)}" data-media-type="${escapeHtml(item.media_type || 'ebook')}" class="px-2.5 py-1 text-xs bg-indigo-600 hover:bg-indigo-500 text-white rounded transition-colors" title="${t('wishlist_search')}">${t('wishlist_search')}</button>
        <button data-action="deleteWishlistItem" data-id="${item.id}" class="px-2.5 py-1 text-xs bg-slate-700 hover:bg-red-600 text-slate-300 hover:text-white rounded transition-colors" title="Remove">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
        </button>
      </div>
    </div>
  `;
}

function showWishlistForm() {
  document.getElementById('wishlist-form').classList.remove('hidden');
  document.getElementById('wl-title').focus();
}

function hideWishlistForm() {
  document.getElementById('wishlist-form').classList.add('hidden');
  document.getElementById('wl-title').value = '';
  document.getElementById('wl-author').value = '';
}

async function addWishlistItem() {
  const title = document.getElementById('wl-title').value.trim();
  const author = document.getElementById('wl-author').value.trim();
  const mediaType = document.getElementById('wl-type').value;

  if (!title) {
    showToast(t('err_title_required'), 'warning');
    return;
  }

  try {
    await apiJson('/api/wishlist', {
      method: 'POST',
      body: JSON.stringify({ title, author, media_type: mediaType }),
    });
    showToast(t('added_to_wishlist'), 'success');
    hideWishlistForm();
    loadWishlist();
  } catch (err) {
    if (err.message !== 'Unauthorized') showToast(t('failed_add_wishlist'), 'error');
  }
}

async function deleteWishlistItem(id) {
  try {
    await api(`/api/wishlist/${id}`, { method: 'DELETE' });
    showToast(t('removed_from_wishlist'), 'success');
    loadWishlist();
  } catch (err) {
    if (err.message !== 'Unauthorized') showToast(t('failed_delete'), 'error');
  }
}

function searchWishlistItem(title, mediaType) {
  const tabMap = { ebook: 'ebooks', audiobook: 'audiobooks', manga: 'manga' };
  switchTab('search');
  switchSearchTab(tabMap[mediaType] || 'ebooks');
  document.getElementById('search-input').value = title;
  doSearch(title);
}

// ============================================================
// SETTINGS
// ============================================================
async function loadSettings() {
  loadConfig();
  loadSources();
  loadTOTPStatus();
  loadSettingToggles();
  showChangePasswordIfMultiUser();
  if (state.currentRole === 'admin') {
    loadUsers();
    loadInviteCodes();
  }
}

// Show the self-service change-password card only when the user is logged in
// against a DB account. Env-credential single-user installs can't change pw
// here — those creds are sourced from AUTH_USERNAME / AUTH_PASSWORD.
async function showChangePasswordIfMultiUser() {
  try {
    const data = await apiJson('/api/auth/status');
    if (data.authenticated && data.user_id) {
      document.getElementById('change-password-section').classList.remove('hidden');
    }
  } catch (err) {}
}

// Guard with optional chaining: the change-password card is only present for
// DB-backed accounts, so this element is absent on env-credential / no-auth
// installs. Without the guard, the null deref threw at the top level and
// aborted the rest of this script — leaving later `const` declarations (e.g.
// INTEGRATION_FIELDS) uninitialized and breaking the integration Save button.
document.getElementById('change-password-form')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const errEl = document.getElementById('cp-error');
  errEl.classList.add('hidden');

  const current = document.getElementById('cp-current').value;
  const next = document.getElementById('cp-new').value;
  const confirm = document.getElementById('cp-confirm').value;

  if (!current || !next || !confirm) {
    errEl.textContent = 'All three fields are required';
    errEl.classList.remove('hidden');
    return;
  }
  if (next.length < 6) {
    errEl.textContent = 'New password must be at least 6 characters';
    errEl.classList.remove('hidden');
    return;
  }
  if (next !== confirm) {
    errEl.textContent = 'New password and confirmation do not match';
    errEl.classList.remove('hidden');
    return;
  }

  try {
    const res = await apiJson('/api/me/password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ current_password: current, new_password: next }),
    });
    if (res.success) {
      document.getElementById('cp-current').value = '';
      document.getElementById('cp-new').value = '';
      document.getElementById('cp-confirm').value = '';
      showToast('Password updated', 'success');
    } else {
      errEl.textContent = res.error || 'Failed to update password';
      errEl.classList.remove('hidden');
    }
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      errEl.textContent = 'Failed to update password';
      errEl.classList.remove('hidden');
    }
  }
});

// Settings keys backed by an input in the Integrations section, grouped by the
// integration name passed to saveIntegration().
const INTEGRATION_FIELDS = {
  annas:          ['annas_archive_domain', 'annas_archive_secret_key'],
  prowlarr:       ['prowlarr_url', 'prowlarr_api_key'],
  qbittorrent:    ['qb_url', 'qb_user', 'qb_pass'],
  transmission:   ['transmission_url', 'transmission_user', 'transmission_pass', 'torrent_client'],
  sabnzbd:        ['sabnzbd_url', 'sabnzbd_api_key', 'sabnzbd_category'],
  audiobookshelf: ['abs_url', 'abs_token'],
  kavita:         ['kavita_url', 'kavita_user', 'kavita_pass', 'kavita_path_map'],
  komga:          ['komga_url', 'komga_user', 'komga_pass'],
  calibre:        ['calibre_url', 'calibre_library_path'],
};

async function loadSettingToggles() {
  try {
    const data = await apiJson('/api/settings');
    const removeTorrent = document.getElementById('remove-torrent-toggle');
    if (removeTorrent && data.remove_torrent_after_import !== undefined) {
      removeTorrent.checked = data.remove_torrent_after_import;
    }
    const langFilter = document.getElementById('foreign-lang-filter-toggle');
    if (langFilter && data.foreign_lang_filter !== undefined) {
      langFilter.checked = data.foreign_lang_filter;
    }
    // Populate integration inputs. Sensitive values come back as "--------"
    // from the server; the save handler drops that sentinel before writing.
    for (const fields of Object.values(INTEGRATION_FIELDS)) {
      for (const key of fields) {
        const el = document.getElementById(`setting-${key}`);
        if (el && data[key] !== undefined && data[key] !== null) {
          el.value = data[key];
        }
      }
    }
  } catch (err) {}
}

async function saveIntegration(name) {
  const fields = INTEGRATION_FIELDS[name];
  if (!fields) return;
  const payload = {};
  for (const key of fields) {
    const el = document.getElementById(`setting-${key}`);
    if (!el) continue;
    payload[key] = el.value;
  }
  try {
    const res = await apiJson('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (res.success) {
      showToast('Saved. Restart the container for new URLs/credentials to take effect.', 'success');
    } else {
      showToast(res.error || 'Failed to save', 'error');
    }
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      showToast('Failed to save', 'error');
    }
  }
}

async function loadConfig() {
  try {
    const data = await apiJson('/api/config');
    state.config = data;

    const configEl = document.getElementById('config-info');
    const items = [];

    if (data.prowlarr) items.push(configItem('Prowlarr', data.prowlarr.url || t('not_configured')));
    if (data.qbittorrent) items.push(configItem('qBittorrent', data.qbittorrent.url || t('not_configured')));
    if (data.transmission) items.push(configItem('Transmission', data.transmission.url || t('not_configured')));
    if (data.sabnzbd) items.push(configItem('SABnzbd', data.sabnzbd.url || t('not_configured')));
    if (data.kavita_url) items.push(configItem('Kavita', data.kavita_url));
    if (data.audiobookshelf_url) items.push(configItem('Audiobookshelf', data.audiobookshelf_url));

    configEl.innerHTML = items.length > 0 ? items.join('') : `<p class="text-slate-500">${t('no_config_data')}</p>`;
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      document.getElementById('config-info').innerHTML = `<p class="text-red-400">${t('failed_load_config')}</p>`;
    }
  }
}

function configItem(label, value) {
  return `
    <div class="flex items-center justify-between py-2 border-b border-slate-800 last:border-0">
      <span class="text-slate-300">${escapeHtml(label)}</span>
      <span class="text-slate-500 text-xs truncate max-w-[60%] text-right">${escapeHtml(value)}</span>
    </div>
  `;
}

async function loadSources() {
  const container = document.getElementById('sources-list');
  const loadingEl = document.getElementById('sources-loading');

  try {
    const data = await apiJson('/api/sources');
    const sources = Array.isArray(data) ? data : (data.sources || []);
    loadingEl.classList.add('hidden');

    if (sources.length === 0) {
      container.innerHTML = `<p class="text-slate-500 text-sm">${t('no_sources')}</p>`;
      return;
    }

    container.innerHTML = sources.map(s => {
      const enabled = s.enabled !== false;
      const tabLabel = s.search_tab || s.download_type || '';
      return `
        <div class="flex items-center justify-between bg-slate-800/50 rounded-lg px-4 py-2.5">
          <div class="flex items-center gap-3">
            <div class="w-2 h-2 rounded-full ${enabled ? 'bg-emerald-400' : 'bg-slate-600'}"></div>
            <span class="text-sm text-slate-300">${escapeHtml(s.label || s.name || 'Unknown')}</span>
            ${tabLabel ? `<span class="text-xs text-slate-600">${escapeHtml(tabLabel)}</span>` : ''}
          </div>
          <span class="text-xs ${enabled ? 'text-emerald-400' : 'text-slate-500'}">${enabled ? t('enabled') : t('disabled')}</span>
        </div>
      `;
    }).join('');
  } catch (err) {
    loadingEl.classList.add('hidden');
    if (err.message !== 'Unauthorized') {
      container.innerHTML = `<p class="text-red-400 text-sm">${t('failed_load_sources')}</p>`;
    }
  }
}

async function toggleForeignLangFilter() {
  const toggle = document.getElementById('foreign-lang-filter-toggle');
  if (!toggle) return;
  const enabled = toggle.checked;
  try {
    const res = await apiJson('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ foreign_lang_filter: enabled })
    });
    if (res.success) {
      showToast(enabled ? t('filter_enabled_toast') : t('filter_disabled_toast'), 'success');
    } else {
      toggle.checked = !enabled;
      showToast(t('filter_update_failed'), 'error');
    }
  } catch (err) {
    toggle.checked = !enabled;
    if (err.message !== 'Unauthorized') {
      showToast(t('filter_save_failed'), 'error');
    }
  }
}

async function toggleRemoveTorrent() {
  const toggle = document.getElementById('remove-torrent-toggle');
  if (!toggle) return;
  const enabled = toggle.checked;
  try {
    const res = await apiJson('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ remove_torrent_after_import: enabled })
    });
    if (res.success) {
      showToast(enabled ? 'Torrents will be removed after import' : 'Torrents will keep seeding after import', 'success');
    } else {
      toggle.checked = !enabled;
      showToast('Failed to update setting', 'error');
    }
  } catch (err) {
    toggle.checked = !enabled;
    if (err.message !== 'Unauthorized') {
      showToast('Failed to save setting', 'error');
    }
  }
}

async function testConnection(service) {
  const statusEl = document.getElementById(`test-${service}-status`);
  statusEl.textContent = t('testing');
  statusEl.className = 'text-xs text-yellow-400';

  try {
    const data = await apiJson(`/api/test/${service}`, { method: 'POST' });
    if (data.success) {
      statusEl.textContent = t('connected');
      statusEl.className = 'text-xs text-emerald-400';
    } else {
      statusEl.textContent = data.error || t('conn_error');
      statusEl.className = 'text-xs text-red-400';
    }
  } catch (err) {
    statusEl.textContent = t('conn_error');
    statusEl.className = 'text-xs text-red-400';
  }
}

// ============================================================
// TOTP SETTINGS
// ============================================================
async function loadTOTPStatus() {
  const section = document.getElementById('totp-settings');
  // The TOTP settings markup is not rendered in every build/mode — without
  // this guard the Settings tab throws an uncaught TypeError on null.
  if (!section) return;
  try {
    const data = await apiJson('/api/totp/status');
    section.classList.remove('hidden');
    if (data.enabled) {
      document.getElementById('totp-disabled-section').classList.add('hidden');
      document.getElementById('totp-enabled-section').classList.remove('hidden');
    } else {
      document.getElementById('totp-disabled-section').classList.remove('hidden');
      document.getElementById('totp-enabled-section').classList.add('hidden');
    }
    document.getElementById('totp-setup-section').classList.add('hidden');
    document.getElementById('totp-disable-section').classList.add('hidden');
  } catch (err) {
    // Not in multi-user mode — hide TOTP settings.
    section.classList.add('hidden');
  }
}

async function setupTOTP() {
  try {
    const data = await apiJson('/api/totp/setup', { method: 'POST' });
    if (!data.success) {
      showToast(data.error || t('failed_setup_totp'), 'error');
      return;
    }
    document.getElementById('totp-disabled-section').classList.add('hidden');
    document.getElementById('totp-setup-section').classList.remove('hidden');
    document.getElementById('totp-secret-display').textContent = data.secret;

    // Generate QR code using a QR API.
    const qrUrl = 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=' + encodeURIComponent(data.qr_url);
    document.getElementById('totp-qr-img').src = qrUrl;

    // Display backup codes.
    const codesEl = document.getElementById('totp-backup-codes');
    codesEl.innerHTML = data.backup_codes.map(c => `<span class="bg-slate-700 px-2 py-1 rounded text-center">${escapeHtml(c)}</span>`).join('');
  } catch (err) {
    showToast(t('failed_setup_totp'), 'error');
  }
}

async function verifyTOTP() {
  const code = document.getElementById('totp-verify-code').value.trim();
  if (!code) {
    showToast(t('enter_6digit_code'), 'warning');
    return;
  }
  try {
    const data = await apiJson('/api/totp/verify', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
    if (data.success) {
      showToast(t('totp_enabled_success'), 'success');
      loadTOTPStatus();
    } else {
      showToast(data.error || t('err_invalid_code'), 'error');
    }
  } catch (err) {
    showToast(t('verification_failed'), 'error');
  }
}

function cancelTOTPSetup() {
  document.getElementById('totp-setup-section').classList.add('hidden');
  document.getElementById('totp-disabled-section').classList.remove('hidden');
}

function showDisableTOTP() {
  document.getElementById('totp-enabled-section').classList.add('hidden');
  document.getElementById('totp-disable-section').classList.remove('hidden');
  document.getElementById('totp-disable-code').focus();
}

function cancelDisableTOTP() {
  document.getElementById('totp-disable-section').classList.add('hidden');
  document.getElementById('totp-enabled-section').classList.remove('hidden');
}

async function disableTOTP() {
  const code = document.getElementById('totp-disable-code').value.trim();
  if (!code) {
    showToast(t('enter_totp_code'), 'warning');
    return;
  }
  try {
    const data = await apiJson('/api/totp/disable', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
    if (data.success) {
      showToast(t('totp_disabled_success'), 'success');
      loadTOTPStatus();
    } else {
      showToast(data.error || t('err_invalid_code'), 'error');
    }
  } catch (err) {
    showToast(t('failed_disable_totp'), 'error');
  }
}

// ============================================================
// USER MANAGEMENT (ADMIN)
// ============================================================
async function loadUsers() {
  const section = document.getElementById('user-management');
  try {
    const data = await apiJson('/api/users');
    if (!data.success) return;
    section.classList.remove('hidden');

    const container = document.getElementById('users-list');
    const users = data.users || [];
    container.innerHTML = users.map(u => {
      const roleOptions = `<select data-action-change="changeUserRole" data-id="${u.id}" class="bg-slate-700 text-sm text-slate-300 rounded px-2 py-1 border-0">
        <option value="user" ${u.role === 'user' ? 'selected' : ''}>${t('role_user')}</option>
        <option value="admin" ${u.role === 'admin' ? 'selected' : ''}>${t('role_admin')}</option>
      </select>`;
      const totpBadge = u.totp_enabled ? '<span class="text-xs text-emerald-400 bg-emerald-400/10 px-1.5 py-0.5 rounded">2FA</span>' : '';
      const lastLogin = u.last_login ? new Date(u.last_login).toLocaleDateString() : t('never');
      return `
        <div class="flex items-center justify-between bg-slate-800/50 rounded-lg px-4 py-3">
          <div class="flex items-center gap-3">
            <span class="text-sm font-medium text-white">${escapeHtml(u.username)}</span>
            ${totpBadge}
            <span class="text-xs text-slate-600">${t('last_login', {date: lastLogin})}</span>
          </div>
          <div class="flex items-center gap-2">
            ${roleOptions}
            <button data-action="deleteUser" data-id="${u.id}" data-username="${escapeHtml(u.username)}" class="px-2 py-1 text-xs bg-slate-700 hover:bg-red-600 text-slate-400 hover:text-white rounded transition-colors" title="${t('failed_delete_user')}">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
            </button>
          </div>
        </div>
      `;
    }).join('');
  } catch (err) {
    // Not admin or multi-user not active — hide.
    section.classList.add('hidden');
  }
}

async function changeUserRole(id, role) {
  try {
    const data = await apiJson(`/api/users/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ role }),
    });
    if (data.success) {
      showToast(t('user_role_updated'), 'success');
    } else {
      showToast(data.error || t('failed_update_role'), 'error');
      loadUsers();
    }
  } catch (err) {
    showToast(t('failed_update_role'), 'error');
    loadUsers();
  }
}

async function deleteUser(id, username) {
  if (!confirm(t('confirm_delete_user', {username: username}))) return;
  try {
    const data = await apiJson(`/api/users/${id}`, { method: 'DELETE' });
    if (data.success) {
      showToast(t('user_deleted'), 'success');
      loadUsers();
    } else {
      showToast(data.error || t('failed_delete_user'), 'error');
    }
  } catch (err) {
    showToast(t('failed_delete_user'), 'error');
  }
}

async function addUser() {
  const username = document.getElementById('new-user-name').value.trim();
  const password = document.getElementById('new-user-pass').value;
  if (!username || !password) {
    showToast(t('err_credentials_required'), 'warning');
    return;
  }
  try {
    const data = await apiJson('/api/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    if (data.success) {
      showToast(t('user_created'), 'success');
      document.getElementById('new-user-name').value = '';
      document.getElementById('new-user-pass').value = '';
      loadUsers();
    } else {
      showToast(data.error || t('failed_create_user'), 'error');
    }
  } catch (err) {
    showToast(t('failed_create_user'), 'error');
  }
}

// ============================================================
// INVITE CODES (admin)
// ============================================================
async function loadInviteCodes() {
  try {
    const data = await apiJson('/api/invites');
    const list = data.invites || [];
    const el = document.getElementById('invite-codes-list');
    if (list.length === 0) {
      el.innerHTML = `<p class="text-xs text-slate-500">No invite codes yet.</p>`;
      return;
    }
    el.innerHTML = list.map(inv => {
      const expires = inv.expires_at
        ? new Date(inv.expires_at * 1000).toLocaleDateString()
        : 'Never';
      const usesLabel = `${inv.uses} / ${inv.max_uses}`;
      const exhausted = inv.uses >= inv.max_uses;
      return `
        <div class="flex items-center justify-between bg-slate-800/50 rounded-lg px-3 py-2 text-sm">
          <div class="flex items-center gap-3 flex-1 min-w-0">
            <code class="text-indigo-400 font-mono text-xs truncate">${escapeHtml(inv.code)}</code>
            <span class="text-xs text-slate-500">role: ${escapeHtml(inv.role)}</span>
            <span class="text-xs ${exhausted ? 'text-amber-400' : 'text-slate-500'}">uses: ${usesLabel}</span>
            <span class="text-xs text-slate-500">expires: ${escapeHtml(expires)}</span>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <button data-action="copyInviteCode" data-code="${escapeHtml(inv.code)}" class="px-2 py-1 text-xs bg-slate-700 hover:bg-slate-600 text-slate-300 rounded transition-colors" title="Copy">Copy</button>
            <button data-action="revokeInviteCode" data-id="${inv.id}" class="px-2 py-1 text-xs bg-red-700 hover:bg-red-600 text-white rounded transition-colors" title="Revoke">Revoke</button>
          </div>
        </div>`;
    }).join('');
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      document.getElementById('invite-codes-list').innerHTML = `<p class="text-xs text-red-400">Failed to load invite codes</p>`;
    }
  }
}

async function generateInviteCode() {
  const role = document.getElementById('invite-role').value;
  const maxUses = parseInt(document.getElementById('invite-max-uses').value, 10) || 1;
  const expiresDays = parseInt(document.getElementById('invite-expires-days').value, 10) || 0;
  const expiresIn = expiresDays > 0 ? expiresDays * 86400 : 0;
  try {
    const res = await apiJson('/api/invites', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ role, max_uses: maxUses, expires_in: expiresIn }),
    });
    if (res.success) {
      showToast('Invite code generated', 'success');
      loadInviteCodes();
    } else {
      showToast(res.error || 'Failed to generate', 'error');
    }
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      showToast('Failed to generate invite', 'error');
    }
  }
}

async function copyInviteCode(code) {
  try {
    await navigator.clipboard.writeText(code);
    showToast('Copied', 'success');
  } catch (err) {
    showToast('Copy failed — select the code manually', 'warning');
  }
}

async function revokeInviteCode(id) {
  if (!confirm('Revoke this invite code? Anyone with it will no longer be able to register.')) return;
  try {
    const res = await apiJson(`/api/invites/${id}`, { method: 'DELETE' });
    if (res.success) {
      showToast('Invite revoked', 'success');
      loadInviteCodes();
    } else {
      showToast(res.error || 'Failed to revoke', 'error');
    }
  } catch (err) {
    if (err.message !== 'Unauthorized') {
      showToast('Failed to revoke', 'error');
    }
  }
}

// ============================================================
// STATS
// ============================================================
async function loadStats() {
  try {
    const data = await apiJson('/api/stats');
    const statsEl = document.getElementById('header-stats');
    if (data.total_items !== undefined) {
      statsEl.textContent = t('n_items_in_library', {n: data.total_items});
      statsEl.classList.remove('hidden');
    }
  } catch (err) {
    // Silently ignore
  }
}

// ============================================================
// BROWSE TAB (filesystem library browser)
// ============================================================
// Backed by /api/files/*, which reads the disk directly — unlike the Library
// tab, nothing here goes through Kavita. One directory per request, so the UI
// never waits on a recursive walk of an NFS share.
//
// Rows are built with DOM APIs rather than HTML strings: file names come from
// the filesystem and may contain quotes, which escapeHtml() does not escape and
// which would break out of an attribute.

const BROWSE_FOLDER_ICON = '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>';
const BROWSE_FILE_ICON = '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 3h7l5 5v13a1 1 0 01-1 1H7a1 1 0 01-1-1V4a1 1 0 011-1z"/></svg>';
const BROWSE_PENCIL_ICON = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>';
const BROWSE_TAG_ICON = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V7a4 4 0 014-4z"/></svg>';
const BROWSE_UP_ICON ='<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12"/></svg>';

// Kinds this build knows how to label; anything else falls back to "other".
const BROWSE_KINDS = new Set(['cbz', 'epub', 'pdf', 'cbr', 'audio', 'other']);

// Library names are translated from the stable root id; the API's label is only
// a fallback for an id this build does not know about.
function browseRootLabel(root) {
  const key = `browse_root_${root.id}`;
  const label = t(key);
  return label === key ? (root.label || root.id) : label;
}

async function loadBrowse() {
  hideBrowseError();
  try {
    const data = await apiJson('/api/files/roots');
    state.browseRoots = data.roots || [];
    state.browseKavita = !!data.kavita;
  } catch (err) {
    state.browseRoots = [];
    state.browseKavita = false;
    showBrowseError(t('browse_roots_failed'));
  }

  renderBrowseRoots();
  // No Kavita configured means the endpoint can only answer 503, so the button
  // is not offered at all rather than shown and failing.
  document.getElementById('browse-scan-btn').classList
    .toggle('hidden', !state.browseKavita || state.browseRoots.length === 0);
  document.getElementById('browse-no-roots').classList.toggle('hidden', state.browseRoots.length > 0);
  if (state.browseRoots.length === 0) {
    clearBrowseListing();
    return;
  }

  // Keep the folder we were in across tab switches, as long as its library is
  // still configured.
  const known = state.browseRoots.some(r => r.id === state.browseRootId);
  if (!known || !state.browsePath) {
    state.browseRootId = state.browseRoots[0].id;
    state.browsePath = state.browseRoots[0].path;
  }
  await openBrowsePath(state.browsePath);
}

function selectBrowseRoot(id) {
  const root = state.browseRoots.find(r => r.id === id);
  if (!root) return;
  state.browseRootId = root.id;
  openBrowsePath(root.path);
}

async function openBrowsePath(path) {
  hideBrowseError();
  let data;
  try {
    data = await apiJson(`/api/files/list?path=${encodeURIComponent(path)}`);
  } catch (err) {
    // A folder that vanished (renamed, deleted) must not leave the tab stuck on
    // a dead path: fall back to the library root.
    const root = state.browseRoots.find(r => r.id === state.browseRootId);
    if (root && path !== root.path) {
      state.browsePath = root.path;
      await openBrowsePath(root.path);
    }
    showBrowseError(t('browse_load_failed'));
    return;
  }

  state.browseRootId = data.root;
  state.browsePath = data.path;
  state.browseIsRoot = !!data.is_root;
  // Leaving the folder invalidates the plan that was previewed on it.
  if (state.renamePath && state.renamePath !== data.path) closeBrowseRename();
  if (state.metaPath && state.metaPath !== data.path) closeBrowseMeta();
  // The file panel is anchored to one file, which the new listing no longer shows.
  if (state.metaFilePath) closeBrowseFileMeta();
  // A library root is never renamable, so the button has nothing to offer there.
  document.getElementById('browse-rename-btn').classList.toggle('hidden', state.browseIsRoot);
  // Same for a folder-wide metadata edit: at the root it would cover every series.
  document.getElementById('browse-meta-btn').classList.toggle('hidden', state.browseIsRoot);
  renderBrowseRoots();
  renderBrowseBreadcrumb(data);
  renderBrowseEntries(data);
}

function renderBrowseRoots() {
  const wrap = document.getElementById('browse-roots');
  wrap.innerHTML = '';
  state.browseRoots.forEach(root => {
    const btn = document.createElement('button');
    btn.className = 'px-4 py-1.5 text-sm font-medium rounded-lg transition-colors whitespace-nowrap ' +
      (root.id === state.browseRootId ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white');
    btn.dataset.action = 'browseRoot';
    btn.dataset.arg = root.id;
    btn.textContent = browseRootLabel(root);
    wrap.appendChild(btn);
  });
}

function renderBrowseBreadcrumb(data) {
  const nav = document.getElementById('browse-breadcrumb');
  nav.innerHTML = '';
  const segments = data.segments || [];
  const rootMeta = state.browseRoots.find(r => r.id === data.root)
    || { id: data.root, label: data.root_label };
  nav.appendChild(browseCrumb(browseRootLabel(rootMeta), data.root_path, segments.length === 0));
  segments.forEach((seg, i) => {
    const sep = document.createElement('span');
    sep.className = 'text-slate-600';
    sep.textContent = '/';
    nav.appendChild(sep);
    nav.appendChild(browseCrumb(seg.name, seg.path, i === segments.length - 1));
  });
}

function browseCrumb(label, path, isCurrent) {
  if (isCurrent) {
    const span = document.createElement('span');
    span.className = 'text-slate-200 font-medium';
    span.textContent = label;
    return span;
  }
  const btn = document.createElement('button');
  btn.className = 'text-indigo-400 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'browseOpen';
  btn.dataset.path = path;
  btn.textContent = label;
  return btn;
}

function clearBrowseListing() {
  closeBrowseRename();
  closeBrowseMeta();
  closeBrowseFileMeta();
  document.getElementById('browse-rename-btn').classList.add('hidden');
  document.getElementById('browse-meta-btn').classList.add('hidden');
  document.getElementById('browse-breadcrumb').innerHTML = '';
  document.getElementById('browse-list').innerHTML = '';
  document.getElementById('browse-empty').classList.add('hidden');
}

function renderBrowseEntries(data) {
  const list = document.getElementById('browse-list');
  list.innerHTML = '';
  const entries = data.entries || [];
  document.getElementById('browse-empty').classList.toggle('hidden', entries.length > 0);

  if (data.parent) {
    list.appendChild(browseRow({ name: '..', is_dir: true }, data.parent, true));
  }
  entries.forEach(entry => {
    list.appendChild(browseRow(entry, browseJoin(data.path, entry.name), false));
  });
}

// The API returns POSIX paths (the server runs on Linux); a child path is the
// directory plus its entry name.
function browseJoin(dir, name) {
  return dir.endsWith('/') ? dir + name : `${dir}/${name}`;
}

function browseRow(entry, targetPath, isParent) {
  const clickable = entry.is_dir;
  const row = document.createElement(clickable ? 'button' : 'div');
  row.className = 'w-full flex items-center gap-3 px-4 py-2.5 text-left'
    + (clickable ? ' hover:bg-slate-800/60 transition-colors' : '');
  if (clickable) {
    row.dataset.action = 'browseOpen';
    row.dataset.path = targetPath;
  }
  if (!isParent && !entry.is_dir) {
    // Anchors the inline rename editor to this row.
    row.dataset.filePath = targetPath;
    row.dataset.fileName = entry.name;
  }

  const icon = document.createElement('span');
  icon.className = 'shrink-0 ' + (entry.is_dir ? 'text-amber-400/80' : 'text-slate-500');
  icon.innerHTML = isParent ? BROWSE_UP_ICON : (entry.is_dir ? BROWSE_FOLDER_ICON : BROWSE_FILE_ICON);
  row.appendChild(icon);

  const name = document.createElement('span');
  name.className = 'browse-row-name flex-1 min-w-0 truncate text-sm ' + (entry.is_dir ? 'text-slate-100' : 'text-slate-300');
  name.textContent = isParent ? t('browse_parent') : entry.name;
  row.appendChild(name);

  if (isParent) return row;

  const meta = document.createElement('span');
  meta.className = 'shrink-0 flex items-center gap-3 text-xs text-slate-500';
  if (!entry.is_dir) {
    meta.appendChild(browseFileMetaButton(targetPath));
    meta.appendChild(browseRenameFileButton());
    meta.appendChild(browseKindBadge(entry));
    const size = document.createElement('span');
    size.className = 'hidden sm:inline tabular-nums';
    size.textContent = formatSize(entry.size || 0);
    meta.appendChild(size);
  }
  const modified = document.createElement('span');
  modified.className = 'hidden md:inline tabular-nums';
  modified.textContent = formatBrowseDate(entry.modified);
  meta.appendChild(modified);
  row.appendChild(meta);

  return row;
}

// Per-file rename trigger: free-form naming for the one file, next to the
// folder-wide plan the Rename panel offers.
function browseRenameFileButton() {
  const btn = document.createElement('button');
  btn.className = 'text-slate-500 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'startBrowseFileRename';
  btn.title = t('browse_rename_file');
  btn.setAttribute('aria-label', t('browse_rename_file'));
  btn.innerHTML = BROWSE_PENCIL_ICON;
  return btn;
}

// Opens the per-file metadata panel. Present on every file, editable or not:
// decision 11 wants a locked format to explain itself, which it can only do
// once the panel is open.
function browseFileMetaButton(path) {
  const btn = document.createElement('button');
  btn.className = 'text-slate-500 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'openBrowseFileMeta';
  btn.dataset.path = path;
  btn.title = t('browse_meta_file');
  btn.setAttribute('aria-label', t('browse_meta_file'));
  btn.innerHTML = BROWSE_TAG_ICON;
  return btn;
}

function browseKindBadge(entry) {
  const badge = document.createElement('span');
  const kind = BROWSE_KINDS.has(entry.kind) ? entry.kind : 'other';
  // Editable formats (CBZ, EPUB) are highlighted; the others carry the
  // rename-only badge, since renaming is the only lever they leave (decision 11).
  badge.className = 'px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide '
    + (entry.editable
      ? 'bg-indigo-500/15 text-indigo-300 border border-indigo-500/30'
      : 'bg-slate-700/40 text-slate-400 border border-slate-600/40');
  badge.textContent = t(`browse_kind_${kind}`);
  if (!entry.editable) badge.title = t('browse_meta_status_unsupported');
  return badge;
}

function formatBrowseDate(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  return isNaN(d.getTime()) ? '' : d.toLocaleDateString();
}

function showBrowseError(message) {
  const el = document.getElementById('browse-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseError() {
  document.getElementById('browse-error').classList.add('hidden');
}

// scanBrowseKavita asks Kavita to re-read the files. Manual by design: renaming
// and metadata edits never trigger it, so a batch of fixes costs one scan
// instead of one per edit.
async function scanBrowseKavita() {
  if (!state.browseRootId) return;
  const btn = document.getElementById('browse-scan-btn');
  const label = btn.querySelector('span');
  const previous = label.textContent;
  btn.disabled = true;
  label.textContent = t('browse_scan_running');
  try {
    await postFiles('/api/files/kavita/scan', { root: state.browseRootId });
    showToast(t('browse_scan_done'), 'success');
  } catch (err) {
    showToast(err.message || t('browse_scan_failed'), 'error');
  } finally {
    btn.disabled = false;
    label.textContent = previous;
  }
}

// ============================================================
// BROWSE — RENAME (folder with preview, and single files)
// ============================================================
// Nothing is ever applied implicitly: the panel previews, the user reads, and
// only the Apply button writes. Every input that feeds the plan (series name,
// scope, per-file volume) re-previews, so what is on screen is always what the
// server would do — the plan is recomputed server-side on apply anyway.

// postFiles POSTs JSON and surfaces the server's error message, which the
// generic apiJson swallows in favour of the status code.
async function postFiles(path, body) {
  const resp = await api(path, { method: 'POST', body: JSON.stringify(body) });
  let data = null;
  try {
    data = await resp.json();
  } catch (err) {
    data = null;
  }
  if (!resp.ok) throw new Error((data && data.error) || `API error: ${resp.status}`);
  return data;
}

function browseBaseName(path) {
  const parts = String(path || '').split('/').filter(Boolean);
  return parts.length ? parts[parts.length - 1] : '';
}

function openBrowseRename() {
  if (!state.browsePath || state.browseIsRoot) return;
  state.renamePath = state.browsePath;
  state.renameVolumes = {};
  state.renamePreview = null;
  // Decision 5: the folder's own name is the starting point, and it is editable.
  document.getElementById('browse-rename-series').value = browseBaseName(state.browsePath);
  document.getElementById('browse-rename-mode').value = 'folder_and_files';
  document.getElementById('browse-rename-panel').classList.remove('hidden');
  refreshBrowseRenamePreview();
}

function closeBrowseRename() {
  state.renamePath = null;
  state.renamePreview = null;
  state.renameVolumes = {};
  document.getElementById('browse-rename-panel').classList.add('hidden');
  hideBrowseRenameError();
}

async function refreshBrowseRenamePreview() {
  if (!state.renamePath) return;
  hideBrowseRenameError();
  let data;
  try {
    data = await postFiles('/api/files/rename/preview', browseRenamePayload());
  } catch (err) {
    state.renamePreview = null;
    showBrowseRenameError(err.message || t('browse_rename_preview_failed'));
    document.getElementById('browse-rename-apply').disabled = true;
    return;
  }
  state.renamePreview = data;
  renderBrowseRenamePreview(data);
}

function browseRenamePayload() {
  return {
    path: state.renamePath,
    series_name: document.getElementById('browse-rename-series').value,
    mode: document.getElementById('browse-rename-mode').value,
    volumes: state.renameVolumes,
  };
}

function renderBrowseRenamePreview(data) {
  const folder = document.getElementById('browse-rename-folder');
  folder.textContent = data.folder_to
    ? t('browse_rename_folder_to', {
      from: browseBaseName(data.folder_from),
      to: browseBaseName(data.folder_to),
    })
    : t('browse_rename_folder_same');

  const warn = document.getElementById('browse-rename-warnings');
  const warnings = data.warnings || [];
  warn.textContent = warnings.join(' · ');
  warn.classList.toggle('hidden', warnings.length === 0);

  const body = document.getElementById('browse-rename-rows');
  body.innerHTML = '';
  const files = data.files || [];
  files.forEach(row => body.appendChild(browseRenameRow(row)));
  if (files.length === 0) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 3;
    td.className = 'px-4 py-3 text-sm text-slate-500';
    td.textContent = t('browse_rename_empty');
    tr.appendChild(td);
    body.appendChild(tr);
  }

  const planned = files.filter(f => f.status === 'rename').length;
  const total = planned + (data.folder_to ? 1 : 0);
  document.getElementById('browse-rename-summary').textContent =
    total === 0 ? t('browse_rename_nothing') : t('browse_rename_summary', { count: planned });
  document.getElementById('browse-rename-apply').disabled = total === 0;
}

function browseRenameRow(row) {
  const tr = document.createElement('tr');

  const current = document.createElement('td');
  current.className = 'px-4 py-2 align-middle';
  const name = document.createElement('div');
  name.className = 'text-slate-300 break-all';
  name.textContent = row.name;
  current.appendChild(name);
  tr.appendChild(current);

  // The detected volume is pre-filled and editable: detection both misses
  // volumes and invents them, so the user must be able to correct it. An empty
  // field means "leave this file alone" (decision 6).
  const volCell = document.createElement('td');
  volCell.className = 'px-4 py-2 align-middle';
  const input = document.createElement('input');
  input.type = 'number';
  input.min = '1';
  input.className = 'w-20 px-2 py-1 bg-slate-800 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-indigo-500';
  input.value = row.volume ? String(row.volume) : '';
  input.placeholder = t('browse_rename_volume_placeholder');
  input.dataset.actionChange = 'browseRenameVolume';
  input.dataset.name = row.name;
  volCell.appendChild(input);
  tr.appendChild(volCell);

  const target = document.createElement('td');
  target.className = 'px-4 py-2 align-middle';
  if (row.status === 'rename') {
    const to = document.createElement('div');
    to.className = 'text-emerald-300 break-all';
    to.textContent = row.target;
    target.appendChild(to);
  } else {
    const why = document.createElement('div');
    why.className = 'text-xs text-slate-500';
    why.textContent = t(`browse_rename_status_${row.status}`);
    target.appendChild(why);
  }
  tr.appendChild(target);

  return tr;
}

async function applyBrowseRename() {
  if (!state.renamePath) return;
  const btn = document.getElementById('browse-rename-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_rename_applying');

  let data;
  try {
    data = await postFiles('/api/files/rename/apply', browseRenamePayload());
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseRenameError(err.message || t('browse_rename_failed'));
    return;
  }
  btn.textContent = label;

  if (data.errors && data.errors.length) {
    showBrowseRenameError(data.errors.join(' · '));
  }
  showToast(t('browse_rename_done', { count: data.renamed || 0 }),
    data.errors && data.errors.length ? 'error' : 'success');

  closeBrowseRename();
  // The folder may have moved: re-open where it actually is now.
  await openBrowsePath(data.path || state.browsePath);
}

function browseRenameVolume(el) {
  const raw = el.value.trim();
  const n = parseInt(raw, 10);
  if (raw === '' || !(n > 0)) {
    delete state.renameVolumes[el.dataset.name];
  } else {
    state.renameVolumes[el.dataset.name] = n;
  }
  refreshBrowseRenamePreview();
}

// --- Single file rename -------------------------------------------------
// An inline editor inside the listing row, so renaming one file needs no modal
// and no plan. The extension is not shown: the server always keeps the original
// one, and offering to edit it would suggest otherwise.
function startBrowseFileRename(el) {
  const row = el.closest('[data-file-path]');
  if (!row || row.querySelector('.browse-rename-inline')) return;
  const path = row.dataset.filePath;
  const fullName = row.dataset.fileName || browseBaseName(path);
  const dot = fullName.lastIndexOf('.');
  const stem = dot > 0 ? fullName.slice(0, dot) : fullName;

  const editor = document.createElement('div');
  editor.className = 'browse-rename-inline flex items-center gap-2 flex-1 min-w-0';

  const input = document.createElement('input');
  input.type = 'text';
  input.value = stem;
  input.className = 'flex-1 min-w-0 px-2 py-1 bg-slate-800 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-indigo-500';
  input.addEventListener('keydown', e => {
    if (e.key === 'Enter') commitBrowseFileRename(path, input.value);
    if (e.key === 'Escape') openBrowsePath(state.browsePath);
  });

  const save = document.createElement('button');
  save.className = 'px-2 py-1 text-xs bg-indigo-600 hover:bg-indigo-500 text-white rounded';
  save.textContent = t('browse_rename_file_save');
  save.dataset.action = 'commitBrowseFileRename';
  save.dataset.path = path;

  const cancel = document.createElement('button');
  cancel.className = 'px-2 py-1 text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 rounded';
  cancel.textContent = t('browse_rename_file_cancel');
  cancel.dataset.action = 'cancelBrowseFileRename';

  editor.appendChild(input);
  editor.appendChild(save);
  editor.appendChild(cancel);

  // Replace the name cell for the duration of the edit.
  const nameEl = row.querySelector('.browse-row-name');
  if (nameEl) nameEl.classList.add('hidden');
  row.insertBefore(editor, nameEl ? nameEl.nextSibling : null);
  input.focus();
  input.select();
}

async function commitBrowseFileRename(path, newName) {
  try {
    await postFiles('/api/files/rename/file', { path, new_name: newName });
  } catch (err) {
    showToast(err.message || t('browse_rename_file_failed'), 'error');
    return;
  }
  showToast(t('browse_rename_file_done'), 'success');
  await openBrowsePath(state.browsePath);
}

function showBrowseRenameError(message) {
  const el = document.getElementById('browse-rename-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseRenameError() {
  document.getElementById('browse-rename-error').classList.add('hidden');
}

// ============================================================
// BROWSE — METADATA (folder propagation and single file)
// ============================================================
// Two panels, two different contracts.
//
// The folder panel only offers the fields that describe the *work* — a series,
// its publisher, its genres. Title, volume number and ISBN identify one volume,
// and writing the same one into every file of a folder produces a series whose
// volumes all claim to be the same book. The server refuses them on a folder
// too: a guard that lives only here is not a guard.
//
// Nothing is applied implicitly. The folder panel previews first, the preview
// shows the value each field currently holds next to the one that would replace
// it (decision 4: we overwrite, but never without showing), and Apply is a
// separate click. Editing any input drops the preview, so Apply can never send
// a plan the user is no longer looking at.
//
// The file panel is pre-filled from the file itself, so what would be
// overwritten is already on screen; it saves directly.

// Fields that describe the work and therefore propagate over a folder.
const BROWSE_META_FOLDER_FIELDS = ['series', 'summary', 'publisher', 'language', 'authors', 'genres', 'tags'];
// Everything the single-file panel offers, per-volume fields included.
const BROWSE_META_FILE_FIELDS = ['title', 'series', 'series_index', 'summary', 'publisher', 'language', 'isbn', 'authors', 'genres', 'tags'];
const BROWSE_META_LIST_FIELDS = new Set(['authors', 'genres', 'tags']);
const BROWSE_META_LONG_FIELDS = new Set(['summary']);

// buildMetaField renders one labelled input. withToggle adds the checkbox that
// arms the field: the API distinguishes "leave this alone" (key absent) from
// "set it to empty" (key present, empty), and only an explicit tick can express
// the difference.
function buildMetaField(field, idPrefix, withToggle) {
  const wrap = document.createElement('label');
  wrap.className = 'block';

  const head = document.createElement('span');
  head.className = 'flex items-center gap-2 mb-1';
  if (withToggle) {
    const cb = document.createElement('input');
    cb.type = 'checkbox';
    cb.id = `${idPrefix}-en-${field}`;
    cb.className = 'rounded border-slate-600 bg-slate-800';
    cb.dataset.actionChange = 'browseMetaToggle';
    cb.dataset.field = field;
    head.appendChild(cb);
  }
  const label = document.createElement('span');
  label.className = 'text-xs text-slate-400';
  label.textContent = t(`browse_meta_f_${field}`)
    + (BROWSE_META_LIST_FIELDS.has(field) ? ` (${t('browse_meta_list_hint')})` : '');
  head.appendChild(label);
  wrap.appendChild(head);

  const input = document.createElement(BROWSE_META_LONG_FIELDS.has(field) ? 'textarea' : 'input');
  if (input.tagName === 'TEXTAREA') input.rows = 3;
  else input.type = 'text';
  input.id = `${idPrefix}-${field}`;
  input.className = 'w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed';
  if (withToggle) {
    // Folder panel only: editing a value invalidates the preview it was
    // measured against. The file panel has no preview to drop.
    input.dataset.actionChange = 'browseMetaDirty';
    input.disabled = true;
  }
  wrap.appendChild(input);

  return wrap;
}

function browseMetaValue(idPrefix, field) {
  const el = document.getElementById(`${idPrefix}-${field}`);
  return el ? el.value : '';
}

// browseMetaPayloadFields turns the panel into the API's `fields` object. Only
// the fields it returns are touched; the rest are left exactly as they are.
function browseMetaPayloadFields(idPrefix, fields, isEnabled) {
  const out = {};
  fields.forEach(field => {
    if (!isEnabled(field)) return;
    const raw = browseMetaValue(idPrefix, field);
    if (BROWSE_META_LIST_FIELDS.has(field)) {
      out[field] = raw.split(',').map(v => v.trim()).filter(Boolean);
    } else {
      out[field] = raw;
    }
  });
  return out;
}

// renderMetaNotices shows the informational side of a response. Notices are
// never errors: "12 files are already up to date" is an answer, not a fault,
// and the two must not share a red box.
function renderMetaNotices(el, notices) {
  const lines = (notices || []).map(n => t(`browse_meta_notice_${n.key}`, { count: n.count }));
  el.textContent = lines.join(' · ');
  el.classList.toggle('hidden', lines.length === 0);
}

// --- Folder panel --------------------------------------------------------

function openBrowseMeta() {
  if (!state.browsePath || state.browseIsRoot) return;
  state.metaPath = state.browsePath;
  state.metaPreview = null;

  const host = document.getElementById('browse-meta-fields');
  host.innerHTML = '';
  BROWSE_META_FOLDER_FIELDS.forEach(field => {
    host.appendChild(buildMetaField(field, 'browse-meta', true));
  });
  document.getElementById('browse-meta-recursive').checked = false;
  document.getElementById('browse-meta-append').checked = false;

  // Online lookup: source routed from the library we are browsing, query
  // seeded with the folder name and editable before searching.
  document.getElementById('browse-online-source').value =
    browseOnlineDefaultType(state.browseRootId);
  document.getElementById('browse-online-query').value = browseBaseName(state.browsePath);
  resetBrowseOnline();

  document.getElementById('browse-meta-panel').classList.remove('hidden');
  resetBrowseMetaPreview();
}

function closeBrowseMeta() {
  state.metaPath = null;
  state.metaPreview = null;
  const panel = document.getElementById('browse-meta-panel');
  if (panel) panel.classList.add('hidden');
  hideBrowseMetaError();
  resetBrowseOnline();
}

function browseMetaEnabled(field) {
  const cb = document.getElementById(`browse-meta-en-${field}`);
  return !!(cb && cb.checked);
}

// Any change to the form invalidates the preview: the Apply button must never
// be armed against a plan that is no longer on screen.
function resetBrowseMetaPreview() {
  state.metaPreview = null;
  document.getElementById('browse-meta-preview').classList.add('hidden');
  document.getElementById('browse-meta-rows').innerHTML = '';
  document.getElementById('browse-meta-notices').classList.add('hidden');
  document.getElementById('browse-meta-summary').textContent = '';
  document.getElementById('browse-meta-apply').disabled = true;
}

function browseMetaToggle(el) {
  const input = document.getElementById(`browse-meta-${el.dataset.field}`);
  if (input) {
    input.disabled = !el.checked;
    if (el.checked) input.focus();
  }
  resetBrowseMetaPreview();
}

function browseMetaPayload() {
  return {
    path: state.metaPath,
    recursive: document.getElementById('browse-meta-recursive').checked,
    mode: document.getElementById('browse-meta-append').checked ? 'append' : 'set',
    fields: browseMetaPayloadFields('browse-meta', BROWSE_META_FOLDER_FIELDS, browseMetaEnabled),
  };
}

async function previewBrowseMeta() {
  if (!state.metaPath) return;
  hideBrowseMetaError();
  const payload = browseMetaPayload();
  if (Object.keys(payload.fields).length === 0) {
    showBrowseMetaError(t('browse_meta_no_fields'));
    return;
  }

  let data;
  try {
    data = await postFiles('/api/files/metadata/preview', payload);
  } catch (err) {
    resetBrowseMetaPreview();
    showBrowseMetaError(err.message || t('browse_meta_preview_failed'));
    return;
  }
  state.metaPreview = data;
  renderBrowseMetaPreview(data);
}

function renderBrowseMetaPreview(data) {
  const body = document.getElementById('browse-meta-rows');
  body.innerHTML = '';
  (data.files || []).forEach(file => browseMetaRows(file).forEach(tr => body.appendChild(tr)));
  document.getElementById('browse-meta-preview').classList.remove('hidden');

  renderMetaNotices(document.getElementById('browse-meta-notices'), data.notices);

  const willWrite = (data.files || []).filter(f => f.status === 'edit').length;
  document.getElementById('browse-meta-summary').textContent =
    willWrite === 0 ? t('browse_meta_none') : t('browse_meta_summary', { count: willWrite });
  document.getElementById('browse-meta-apply').disabled = willWrite === 0;
}

// One <tr> per changed field, plus a single explanatory row for files that hold
// no writable metadata — the reason is what makes the row worth showing.
function browseMetaRows(file) {
  const rows = [];
  if (file.status !== 'edit') {
    rows.push(browseMetaStatusRow(file));
    return rows;
  }
  const changes = (file.changes || []).filter(c => c.changed);
  if (changes.length === 0) {
    rows.push(browseMetaStatusRow(file));
    return rows;
  }
  changes.forEach((change, i) => {
    const tr = document.createElement('tr');
    tr.appendChild(browseMetaCell(i === 0 ? file.name : '', 'text-slate-300 break-all'));
    tr.appendChild(browseMetaCell(t(`browse_meta_f_${change.field}`), 'text-slate-400'));
    tr.appendChild(browseMetaCell(change.current || t('browse_meta_empty_value'),
      change.current ? 'text-amber-300/90 break-all' : 'text-slate-600 break-all'));
    tr.appendChild(browseMetaCell(change.target || t('browse_meta_empty_value'),
      change.target ? 'text-emerald-300 break-all' : 'text-slate-600 break-all'));
    rows.push(tr);
  });
  return rows;
}

function browseMetaStatusRow(file) {
  const tr = document.createElement('tr');
  tr.appendChild(browseMetaCell(file.name, 'text-slate-400 break-all'));
  const td = document.createElement('td');
  td.colSpan = 3;
  td.className = 'px-4 py-2 align-middle text-xs text-slate-500';
  let text = t(`browse_meta_status_${file.status}`);
  if (file.reason) text += ' — ' + t(`browse_meta_reason_${file.reason}`);
  if (file.error) text += ' — ' + file.error;
  td.textContent = text;
  tr.appendChild(td);
  return tr;
}

function browseMetaCell(text, cls) {
  const td = document.createElement('td');
  td.className = 'px-4 py-2 align-middle text-sm ' + cls;
  td.textContent = text;
  return td;
}

async function applyBrowseMeta() {
  // Only a preview arms this button, so applying always follows a look.
  if (!state.metaPath || !state.metaPreview) return;
  const btn = document.getElementById('browse-meta-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_meta_applying');

  let data;
  try {
    data = await postFiles('/api/files/metadata/apply', browseMetaPayload());
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseMetaError(err.message || t('browse_meta_failed'));
    return;
  }
  btn.textContent = label;

  const failed = data.errors && data.errors.length;
  if (failed) showBrowseMetaError(data.errors.join(' · '));
  else hideBrowseMetaError();
  renderMetaNotices(document.getElementById('browse-meta-notices'), data.notices);
  showToast(t('browse_meta_done', { count: data.written || 0 }), failed ? 'error' : 'success');

  resetBrowseMetaPreview();
  await openBrowsePath(state.browsePath);
}

function showBrowseMetaError(message) {
  const el = document.getElementById('browse-meta-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseMetaError() {
  const el = document.getElementById('browse-meta-error');
  if (el) el.classList.add('hidden');
}

// --- Online lookup (folder panel only) -----------------------------------
//
// /api/metadata/search is source-agnostic and knows nothing about Kavita, so
// Browse reuses it as-is. The source is routed from the library we are in
// (manga → AniList, ebooks/audiobooks → Open Library) and stays overridable:
// a light novel filed under ebooks is legitimately an AniList entry.
//
// Picking a result fills the fields it carries *and* ticks them, because a
// value nobody armed would silently not be written. Fields the source did not
// return are left blank and unticked — the panel's own contract. Nothing
// reaches disk without Preview then Apply.
//
// Rows are built with DOM APIs: titles and summaries come from a third party
// and may contain quotes, which escapeHtml() does not escape.

// Last candidate list, kept whole so a click prefills from the object rather
// than from what the DOM managed to carry.
let browseOnlineCandidates = [];

// Which source a library defaults to. Anything that is not manga is a book.
function browseOnlineDefaultType(rootId) {
  return rootId === 'manga' ? 'manga' : 'book';
}

// Only http(s) links are rendered: a cover_url or site_url is remote data, and
// a javascript: URL in an href would execute on click.
function browseOnlineSafeURL(raw) {
  if (!raw) return '';
  try {
    const u = new URL(raw, window.location.origin);
    return (u.protocol === 'http:' || u.protocol === 'https:') ? u.href : '';
  } catch (err) {
    return '';
  }
}

function resetBrowseOnline() {
  const results = document.getElementById('browse-online-results');
  if (results) results.innerHTML = '';
  const status = document.getElementById('browse-online-status');
  if (status) {
    status.textContent = '';
    status.classList.add('hidden');
  }
  browseOnlineCandidates = [];
}

function showBrowseOnlineStatus(message) {
  const status = document.getElementById('browse-online-status');
  status.textContent = message;
  status.classList.remove('hidden');
}

async function searchBrowseOnline() {
  const input = document.getElementById('browse-online-query');
  const q = input ? input.value.trim() : '';
  // An empty query would only earn a 400; don't spend a round trip on it.
  if (!q) return;
  const type = document.getElementById('browse-online-source').value;
  const btn = document.getElementById('browse-online-btn');
  const label = btn.textContent;

  resetBrowseOnline();
  showBrowseOnlineStatus(t('browse_online_searching'));
  btn.disabled = true;
  btn.textContent = t('browse_online_searching');

  let data;
  try {
    // api() rather than apiJson(): the status code is the whole message here.
    // 503 means the source is not configured, 502 that it failed to answer —
    // two different problems, and "error" would name neither.
    const resp = await api(`/api/metadata/search?type=${encodeURIComponent(type)}&q=${encodeURIComponent(q)}`);
    if (resp.status === 503) {
      showBrowseOnlineStatus(t('browse_online_unavailable'));
      return;
    }
    if (resp.status === 502) {
      showBrowseOnlineStatus(t('browse_online_upstream'));
      return;
    }
    if (!resp.ok) {
      showBrowseOnlineStatus(t('browse_online_failed'));
      return;
    }
    data = await resp.json();
  } catch (err) {
    // Unauthorized already opened the login modal; anything else is the network.
    if (err.message !== 'Unauthorized') showBrowseOnlineStatus(t('browse_online_failed'));
    return;
  } finally {
    btn.disabled = false;
    btn.textContent = label;
  }

  browseOnlineCandidates = data.candidates || [];
  if (browseOnlineCandidates.length === 0) {
    showBrowseOnlineStatus(t('browse_online_none'));
    return;
  }
  document.getElementById('browse-online-status').classList.add('hidden');
  const results = document.getElementById('browse-online-results');
  browseOnlineCandidates.forEach((c, i) => results.appendChild(browseOnlineRow(c, i)));
}

function browseOnlineRow(c, index) {
  const row = document.createElement('div');
  row.className = 'browse-online-candidate flex items-start gap-3 p-2 rounded-lg border border-slate-700 hover:border-indigo-500 cursor-pointer transition-colors';
  row.dataset.action = 'pickBrowseOnline';
  row.dataset.idx = String(index);

  const cover = browseOnlineSafeURL(c.cover_url);
  if (cover) {
    const img = document.createElement('img');
    img.src = cover;
    img.alt = '';
    img.loading = 'lazy';
    img.className = 'h-14 w-10 object-cover rounded flex-shrink-0';
    row.appendChild(img);
  } else {
    const ph = document.createElement('div');
    ph.className = 'h-14 w-10 rounded bg-slate-700 flex-shrink-0';
    row.appendChild(ph);
  }

  const body = document.createElement('div');
  body.className = 'min-w-0 flex-1';

  const head = document.createElement('div');
  head.className = 'flex items-center gap-1.5';
  const title = document.createElement('span');
  title.className = 'text-sm font-medium text-white line-clamp-1';
  title.textContent = c.title || t('browse_online_unknown_title');
  head.appendChild(title);
  const site = browseOnlineSafeURL(c.site_url);
  if (site) {
    // Two same-named candidates are told apart by opening the source page.
    const link = document.createElement('a');
    link.href = site;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    link.className = 'text-indigo-400 hover:text-indigo-300 text-xs';
    link.textContent = '↗';
    head.appendChild(link);
  }
  body.appendChild(head);

  const sub = document.createElement('div');
  sub.className = 'text-xs text-slate-400 line-clamp-1';
  sub.textContent = [c.year, (c.authors || []).join(', ')].filter(Boolean).join(' · ');
  body.appendChild(sub);

  const summary = document.createElement('div');
  summary.className = 'text-xs text-slate-500 line-clamp-2 mt-0.5';
  summary.textContent = c.summary || '';
  body.appendChild(summary);

  row.appendChild(body);
  return row;
}

// pickBrowseOnline fills and arms the fields the candidate carries. `title` is
// the *series* here: a folder holds a work, and the server refuses a `title`
// on a folder target anyway. `year` and `cover_url` are deliberately dropped —
// filewriter has nowhere to put them.
function pickBrowseOnline(index) {
  const c = browseOnlineCandidates[index];
  if (!c) return;

  const fill = (field, value) => {
    if (!value) return;
    const input = document.getElementById(`browse-meta-${field}`);
    const cb = document.getElementById(`browse-meta-en-${field}`);
    if (!input || !cb) return;
    cb.checked = true;
    input.disabled = false;
    input.value = value;
  };

  fill('series', c.title || '');
  fill('summary', c.summary || '');
  fill('publisher', c.publisher || '');
  fill('language', c.language || '');
  fill('authors', (c.authors || []).join(', '));
  fill('genres', (c.genres || []).join(', '));
  fill('tags', (c.tags || []).join(', '));

  // Setting values in code fires no change event, so drop the preview here:
  // it was measured against the form as it was a moment ago.
  resetBrowseMetaPreview();

  document.querySelectorAll('.browse-online-candidate').forEach(el => {
    const chosen = +el.dataset.idx === index;
    el.classList.toggle('border-indigo-500', chosen);
    el.classList.toggle('bg-indigo-500/10', chosen);
  });
  showBrowseOnlineStatus(t('browse_online_picked'));
}

// --- Single-file panel ---------------------------------------------------

async function openBrowseFileMeta(path) {
  closeBrowseMeta();
  state.metaFilePath = path;
  state.metaFileEditable = false;

  const panel = document.getElementById('browse-file-panel');
  const host = document.getElementById('browse-file-fields');
  host.innerHTML = '';
  BROWSE_META_FILE_FIELDS.forEach(field => {
    host.appendChild(buildMetaField(field, 'browse-file', false));
  });
  document.getElementById('browse-file-title').textContent =
    t('browse_meta_file_title', { name: browseBaseName(path) });
  panel.classList.remove('hidden');
  hideBrowseFileMetaError();

  // Metadata is read here and nowhere else (decision 9): opening one archive on
  // demand is affordable, opening every archive of a listing is not.
  let data;
  try {
    data = await apiJson(`/api/files/metadata?path=${encodeURIComponent(path)}`);
  } catch (err) {
    showBrowseFileMetaError(t('browse_meta_load_failed'));
    lockBrowseFileMeta('');
    return;
  }
  if (state.metaFilePath !== path) return; // a newer panel won the race

  const fields = data.fields || {};
  BROWSE_META_FILE_FIELDS.forEach(field => {
    const el = document.getElementById(`browse-file-${field}`);
    if (!el) return;
    const value = fields[field];
    el.value = Array.isArray(value) ? value.join(', ') : (value || '');
  });

  if (!data.editable) {
    lockBrowseFileMeta(data.reason || 'format_unsupported');
    return;
  }
  state.metaFileEditable = true;
  document.getElementById('browse-file-locked').classList.add('hidden');
  document.getElementById('browse-file-save').disabled = false;
}

// A locked file keeps its panel — greyed out, with the reason spelled out.
// Decision 11 asks for three distinct explanations, not one generic one,
// because what the user should do next differs: rename a PDF, convert a CBR,
// nothing at all for audio.
function lockBrowseFileMeta(reason) {
  state.metaFileEditable = false;
  BROWSE_META_FILE_FIELDS.forEach(field => {
    const el = document.getElementById(`browse-file-${field}`);
    if (el) el.disabled = true;
  });
  document.getElementById('browse-file-save').disabled = true;
  const banner = document.getElementById('browse-file-locked');
  if (!reason) {
    banner.classList.add('hidden');
    return;
  }
  banner.textContent = t('browse_meta_status_unsupported') + ' — ' + t(`browse_meta_reason_${reason}`);
  banner.classList.remove('hidden');
}

function closeBrowseFileMeta() {
  state.metaFilePath = null;
  state.metaFileEditable = false;
  const panel = document.getElementById('browse-file-panel');
  if (panel) panel.classList.add('hidden');
  hideBrowseFileMetaError();
}

async function applyBrowseFileMeta() {
  if (!state.metaFilePath || !state.metaFileEditable) return;
  const btn = document.getElementById('browse-file-save');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_meta_saving');

  // Every field is sent, emptied ones included: the panel shows the file's
  // current values, so what is on screen is what will be stored.
  const payload = {
    path: state.metaFilePath,
    mode: 'set',
    fields: browseMetaPayloadFields('browse-file', BROWSE_META_FILE_FIELDS, () => true),
  };

  let data;
  try {
    data = await postFiles('/api/files/metadata/apply', payload);
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseFileMetaError(err.message || t('browse_meta_failed'));
    return;
  }
  btn.textContent = label;
  btn.disabled = false;

  const failed = data.errors && data.errors.length;
  if (failed) {
    showBrowseFileMetaError(data.errors.join(' · '));
  } else {
    hideBrowseFileMetaError();
    closeBrowseFileMeta();
  }
  showToast(t('browse_meta_done', { count: data.written || 0 }), failed ? 'error' : 'success');
  await openBrowsePath(state.browsePath);
}

function showBrowseFileMetaError(message) {
  const el = document.getElementById('browse-file-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseFileMetaError() {
  const el = document.getElementById('browse-file-error');
  if (el) el.classList.add('hidden');
}

// ============================================================
// UTILITIES
// ============================================================
function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function escapeJs(str) {
  if (!str) return '';
  return str.replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '\\"').replace(/\n/g, '\\n');
}

// ============================================================
// KEYBOARD SHORTCUTS
// ============================================================
document.addEventListener('keydown', (e) => {
  // Ctrl/Cmd + K → focus search
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault();
    switchTab('search');
    document.getElementById('search-input').focus();
  }

  // Escape → close modals, clear search
  if (e.key === 'Escape') {
    hideWishlistForm();
  }
});

// ============================================================
// INIT
// ============================================================
async function init() {
  try {
    // Test auth by hitting config
    const cfg = await apiJson('/api/config');
    state.config = cfg;

    // Update user header.
    if (cfg.current_user) {
      updateUserHeader(cfg.current_user, cfg.current_role);
    }

    // Auth OK
    loadStats();
    // Show default tab
    document.getElementById('search-empty').classList.remove('hidden');
    // Apply language on load
    applyLanguage();
  } catch (err) {
    if (err.message === 'Unauthorized') {
      // Login modal already shown by api()
    }
  }
}

// Start
init();

// ============================================================
// EVENT DELEGATION
// ============================================================
// Replaces all inline on*= attributes so the UI runs under a strict
// Content-Security-Policy (script-src 'self'). Markup declares intent via
// data-action="..."; this explicit whitelist maps it to code — markup can
// never invoke anything that isn't registered here.
const CLICK_ACTIONS = {
  switchTab: el => switchTab(el.dataset.arg),
  switchSearchTab: el => switchSearchTab(el.dataset.arg),
  switchLibraryTab: el => switchLibraryTab(el.dataset.arg),
  browseRoot: el => selectBrowseRoot(el.dataset.arg),
  browseOpen: el => openBrowsePath(el.dataset.path),
  scanBrowseKavita: () => scanBrowseKavita(),
  openBrowseRename: () => openBrowseRename(),
  closeBrowseRename: () => closeBrowseRename(),
  applyBrowseRename: () => applyBrowseRename(),
  startBrowseFileRename: el => startBrowseFileRename(el),
  commitBrowseFileRename: el => {
    const editor = el.closest('.browse-rename-inline');
    const input = editor && editor.querySelector('input');
    if (input) commitBrowseFileRename(el.dataset.path, input.value);
  },
  // Cancelling just re-renders the listing from the server.
  cancelBrowseFileRename: () => openBrowsePath(state.browsePath),
  openBrowseMeta: () => openBrowseMeta(),
  closeBrowseMeta: () => closeBrowseMeta(),
  previewBrowseMeta: () => previewBrowseMeta(),
  applyBrowseMeta: () => applyBrowseMeta(),
  openBrowseFileMeta: el => openBrowseFileMeta(el.dataset.path),
  searchBrowseOnline: () => searchBrowseOnline(),
  // The candidate's source link sits inside the clickable row: opening it is
  // not choosing the candidate.
  pickBrowseOnline: (el, e) => {
    if (e.target.closest && e.target.closest('a')) return;
    pickBrowseOnline(+el.dataset.idx);
  },
  closeBrowseFileMeta: () => closeBrowseFileMeta(),
  applyBrowseFileMeta: () => applyBrowseFileMeta(),
  setSortMode: el => setSortMode(el.dataset.arg),
  setViewLayout: el => setViewLayout(el.dataset.arg),
  setViewGroup: el => setViewGroup(el.dataset.arg),
  toggleLane: el => toggleLane(el.dataset.arg),
  testConnection: el => testConnection(el.dataset.arg),
  saveIntegration: el => saveIntegration(el.dataset.arg),
  toggleMobileNav: () => toggleMobileNav(),
  toggleLanguage: () => toggleLanguage(),
  showLoginForm: () => showLoginForm(),
  showRegisterForm: () => showRegisterForm(),
  showWishlistForm: () => showWishlistForm(),
  hideWishlistForm: () => hideWishlistForm(),
  addWishlistItem: () => addWishlistItem(),
  generateInviteCode: () => generateInviteCode(),
  doLogout: () => doLogout(),
  clearCompleted: () => clearCompleted(),
  addUser: () => addUser(),
  refreshDownloads: () => refreshDownloads(true), // toolbar button forces a refresh
  // Dynamically rendered rows/cards:
  startDownload: el => {
    // data-key resolves against state.resultsByKey (stable across the multiple
    // result lists that swimlanes render — the old data-idx scheme could not).
    const r = state.resultsByKey.get(el.dataset.key);
    if (r) startDownload(r);
  },
  retryDownload: el => retryDownload(el.dataset.jobId),
  deleteLibraryItem: el => deleteLibraryItem(el.dataset.id, el.dataset.type, el.dataset.title),
  goLibraryPage: el => goLibraryPage(+el.dataset.page),
  toggleSelectMode: () => toggleSelectMode(),
  toggleSeriesSelect: el => toggleSeriesSelect(el.dataset.id, el.dataset.name),
  clearSelection: () => clearSelection(),
  openMetadataModal: () => openMetadataModal(),
  closeMetadataModal: () => closeMetadataModal(),
  saveMetadata: () => saveMetadata(),
  searchOnlineMetadata: () => searchOnlineMetadata(),
  applyMetadataCandidate: el => applyMetadataCandidate(+el.dataset.idx),
  searchWishlistItem: el => searchWishlistItem(el.dataset.title, el.dataset.mediaType),
  deleteWishlistItem: el => deleteWishlistItem(+el.dataset.id),
  deleteUser: el => deleteUser(+el.dataset.id, el.dataset.username),
  copyInviteCode: el => copyInviteCode(el.dataset.code),
  revokeInviteCode: el => revokeInviteCode(+el.dataset.id),
};

document.addEventListener('click', e => {
  const el = e.target.closest('[data-action]');
  if (!el) return;
  const fn = CLICK_ACTIONS[el.dataset.action];
  if (!fn) return;
  // Anchors previously used inline `return false` — keep them from navigating.
  if (el.tagName === 'A') e.preventDefault();
  fn(el, e);
});

const CHANGE_ACTIONS = {
  changeUserRole: el => changeUserRole(+el.dataset.id, el.value),
  browseRenameMode: () => refreshBrowseRenamePreview(),
  browseRenameVolume: el => browseRenameVolume(el),
  browseMetaToggle: el => browseMetaToggle(el),
  // Candidates belong to the source that produced them: switching source
  // leaves a list that no longer matches the selector above it.
  browseOnlineSource: () => resetBrowseOnline(),
  // Any edit to the folder form drops the preview, so Apply always follows a
  // fresh look at what would change.
  browseMetaDirty: () => resetBrowseMetaPreview(),
  toggleForeignLangFilter: () => toggleForeignLangFilter(),
  toggleRemoveTorrent: () => toggleRemoveTorrent(),
};

document.addEventListener('change', e => {
  const el = e.target.closest('[data-action-change]');
  if (el) {
    const fn = CHANGE_ACTIONS[el.dataset.actionChange];
    if (fn) fn(el, e);
    return;
  }
  // Bulk-metadata row toggles: enabling the checkbox unlocks the value input.
  if (e.target.classList && e.target.classList.contains('meta-enable')) {
    const row = e.target.closest('.metadata-row');
    if (row) {
      const val = row.querySelector('.meta-value');
      val.disabled = !e.target.checked;
      if (e.target.checked) val.focus();
      // Toggling Series changes what a rename would produce.
      if (row.dataset.field === 'series') refreshRenamePreview();
    }
  }
  // Rename controls: the preview must follow every input that feeds it.
  if (e.target.id === 'metadata-rename-enable') {
    document.getElementById('metadata-rename-options')
      .classList.toggle('hidden', !e.target.checked);
    refreshRenamePreview();
  }
  if (e.target.name === 'meta-rename-mode') refreshRenamePreview();
});

// Editing the series name re-previews, debounced so typing doesn't hammer the
// endpoint (each call hits Kavita once per selected series).
let renamePreviewTimer = null;
document.addEventListener('input', e => {
  const row = e.target.closest && e.target.closest('.metadata-row[data-field="series"]');
  if (!row) return;
  clearTimeout(renamePreviewTimer);
  renamePreviewTimer = setTimeout(refreshRenamePreview, 400);
});

// Same idea for the Browse rename panel. Only the rows are re-rendered, so the
// series field keeps its focus and caret while the preview refreshes under it.
let browseRenameTimer = null;
document.addEventListener('input', e => {
  if (e.target.id !== 'browse-rename-series') return;
  clearTimeout(browseRenameTimer);
  browseRenameTimer = setTimeout(refreshBrowseRenamePreview, 400);
});

// Cover-image fallback (replaces inline onerror=). 'error' events don't
// bubble, so listen in the capture phase.
document.addEventListener('error', e => {
  const img = e.target;
  if (img instanceof HTMLImageElement && img.dataset.phTitle !== undefined) {
    const idx = +(img.dataset.phIdx || 0);
    img.outerHTML = img.dataset.phSize === 'sm'
      ? makeSmallPlaceholderHtml(img.dataset.phTitle, idx)
      : window.makePlaceholder(img.dataset.phTitle, idx);
  }
}, true);
