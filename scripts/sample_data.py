#!/usr/bin/python3

# Setup:
# pip install httpx tqdm pyqt6

import argparse
import base64
import random
import signal
import string
from datetime import datetime, timedelta
from typing import List, Union, Callable

import httpx
from httpx import RequestError

signal.signal(signal.SIGINT, signal.SIG_DFL)  # allow to be closed with sigint, see https://stackoverflow.com/a/6072360/3112139

MACHINE = "devmachine"

PROJECTS = [
    'web-dashboard', 'auth-service', 'payment-gateway', 'user-api',
    'notification-service', 'data-pipeline', 'analytics-platform',
    'admin-portal', 'search-engine', 'recommendation-engine',
    'content-cms', 'booking-system', 'inventory-tracker', 'chat-server',
    'file-uploader', 'ml-training', 'etl-jobs', 'ci-runner',
    'k8s-operator', 'graphql-bff', 'edge-functions', 'billing-worker',
    'session-cache', 'feature-flags', 'log-shipper', 'metrics-collector',
    'trace-aggregator', 'pdf-renderer', 'ocr-pipeline', 'release-bot',
]

LANGUAGES = {
    'Go': 'go',
    'Java': 'java',
    'JavaScript': 'js',
    'TypeScript': 'ts',
    'Python': 'py',
    'PHP': 'php',
    'Ruby': 'rb',
    'Rust': 'rs',
    'C': 'c',
    'C++': 'cpp',
    'C#': 'cs',
    'Swift': 'swift',
    'Kotlin': 'kt',
    'Scala': 'scala',
    'Dart': 'dart',
    'R': 'r',
    'Haskell': 'hs',
    'Lua': 'lua',
    'Elixir': 'ex',
    'Clojure': 'clj',
    'Shell': 'sh',
    'HTML': 'html',
    'CSS': 'css',
    'SCSS': 'scss',
    'Vue': 'vue',
    'Svelte': 'svelte',
    'SQL': 'sql',
    'YAML': 'yaml',
    'Blade': 'blade.php',  # https://github.com/muety/wakapi/issues/172
    '?': 'astro',  # simulate language unknown to wakatime-cli
}

BRANCHES = ['main', 'master', 'develop', 'feature-1', 'feature-2']

# Types follow wakatime-cli/pkg/heartbeat/entity.go (file, domain, url, event, app).
# Categories follow wakatime-cli/pkg/heartbeat/category.go.
TYPE_CATEGORIES = {
    'file': [
        '',  # most file heartbeats arrive without an explicit category
        'coding', 'debugging', 'building', 'code reviewing', 'writing tests', 'writing docs', 'ai coding', 'indexing', 'planning', 'learning', 'designing',
    ],
    'domain': ['browsing', 'researching', 'learning', 'communicating'],
    'url': ['browsing', 'researching', 'learning', 'communicating', 'supporting'],
    'app': ['communicating', 'meeting', 'designing', 'building'],
    'event': ['meeting', 'planning', 'advising'],
}

# Weighted type selection ('file' dominates real-world traffic)
TYPE_WEIGHTS = [('file', 78), ('domain', 8), ('url', 5), ('app', 5), ('event', 4)]

DOMAINS = [
    'github.com', 'stackoverflow.com', 'docs.python.org', 'developer.mozilla.org',
    'pkg.go.dev', 'npmjs.com', 'crates.io', 'rubygems.org', 'mvnrepository.com',
    'dev.to', 'medium.com', 'hackerrank.com', 'leetcode.com', 'figma.com',
    'notion.so', 'confluence.atlassian.com', 'jira.atlassian.com',
]
URLS = [
    'https://github.com/{org}/{repo}/pull/123', 'https://github.com/{org}/{repo}/issues/42',
    'https://stackoverflow.com/questions/12345/how-to-parse-json-in-python',
    'https://docs.python.org/3/library/json.html', 'https://pkg.go.dev/encoding/json',
    'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array',
    'https://leetcode.com/problems/two-sum/', 'https://www.npmjs.com/package/express',
    'https://crates.io/crates/serde', 'https://dev.to/someauthor/building-rest-apis-with-go',
    'https://figma.com/file/abc123/design-spec', 'https://notion.so/workspace/doc/456',
]
APPS = ['Slack', 'Discord', 'Figma', 'Terminal', 'Docker Desktop', 'Notion', 'Postman', 'Insomnia']
EVENTS = [
    'Daily Standup', 'Sprint Planning', 'Backlog Grooming', '1:1 with Manager',
    'Architecture Review', 'Retrospective', 'Pair Programming', 'Code Review Session',
]

USER_AGENTS = [
    # --- plain IDE editors ---
    'wakatime/13.0.7 (Linux-4.15.0-96-generic-x86_64-with-glibc2.4) Python3.8.0.final.0 GoLand/2019.3.4 GoLand-wakatime/11.0.1',
    'wakatime/13.0.4 (Linux-5.4.64-x86_64-with-glibc2.2.5) Python3.7.6.final.0 emacs-wakatime/1.0.2',
    'wakatime/v1.18.11 (linux-5.13.8-200.fc34.x86_64-x86_64) go1.16.7 emacs-wakatime/1.0.2',
    'wakatime/unset (linux-5.11.0-44-generic-x86_64) go1.16.13 emacs-wakatime/1.0.2',
    'wakatime/ (Linux-6.0.42-1-lts-foobar-x86_64) KTextEditor/5.111.0 kate-wakatime/1.3.10',
    'wakatime/v1.86.5 (linux-6.6.4-200.fc39.x86_64-unknown) go1.21.3 neovim/900 vim-wakatime/11.1.1',
    'wakatime/v1.102.1 (windows-10.0.27723.1000-x86_64) go1.22.5 Skype/unknown windows-wakatime/0.5.0',
    'wakatime/v1.102.1 (windows-10.0.27718.1000-x86_64) go1.22.5 Notepad++/unknown windows-wakatime/0.5.0',
    'wakatime/v1.105.0 (linux-6.11.9-zen1-1-zen-unknown) go1.23.3 vscode/1.95.3 vscode-wakatime/24.8.0',
    'wakatime/v1.105.0 (linux-6.11.8-zen1-2-zen-unknown) go1.23.3 cursor/1.93.1 vscode-wakatime/24.8.0',
    'wakatime/v1.106.1 (linux-5.15.167.4-microsoft-standard-WSL2-unknown) go1.23.3 cursor/1.93.1 vscode-wakatime/24.9.2',
    'wakatime/v1.115.2 (windows-10.0.22631.5335-x86_64) go1.24.2 vscode/1.95.3 vscode-wakatime/24.8.0',
    'wakatime/v1.123.0 (darwin-23.4.0-arm64) go1.24.4 windsurf/1.99.3 vscode-wakatime/25.1.1',
    'wakatime/v1.124.1 (windows-10.0.26100.4652-x86_64) go1.24.4 kiro/1.94.0 vscode-wakatime/25.2.0',
    'wakatime/v1.131.0 (darwin-25.0.0-arm64) go1.24.4 vscode/1.95.3 vscode-wakatime/24.8.0',
    'wakatime/v1.139.1 (darwin-25.2.0-arm64) go1.25.5 helix/25.07.1 (74075bb5) wakatime-ls/0.2.2 helix-wakatime/0.2.2',
    'wakatime/1.139.1 (linux-6.18.8-unknown) go1.25.5 helix/25.07.1 (74075bb5) wakatime-ls/0.2.2 helix-wakatime/0.2.2',
    # --- AI coding agents (with model tokens) ---
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 opus/4.1-medium claude-code/2.1.45',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 opus/4.1-medium claude-code/2.1.45 vscode-wakatime/24.8.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 gpt/5.5-high codex-cli/0.141.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 gpt/5 github-copilot-cli/1.2.3 copilot/4.5.6',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 gpt/5.2 opencode-cli/1.0.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 qwen/3-coder-plus qwen-code-cli/1.0.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 gemini/3-flash-preview vscode/1.95.3 vscode-wakatime/24.8.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 gemini/2.0-flash antigravity-cli/1.0.10 antigravity-cli-wakatime/1.0.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 claude-3-5-sonnet/20241022 grok-build/1.0.0',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 composer/2.5 Cursor/1.105.1',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 swe/1.5 Windsurf/1.99.3 vscode-wakatime/25.1.1',
    'wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 M/3.0 claude-code/2.1.45',
    'wakatime/v1.0 (darwin-25.0.0-arm64) go1.21 opus/4.1-medium claude-code/2.1.45',
    'wakatime/v1.0 (windows-10.0.26100.4652-x86_64) go1.21 gpt/5.5-high codex-cli/0.141.0',
    'wakatime/v1.0 (darwin-23.4.0-arm64) go1.21 composer/2.5 Cursor/1.105.1',
    # --- AI harness without a separate model token ---
    'wakatime/v2.7.0 (linux-6.19.12-200.fc43.x86_64-unknown) go1.25.9 Claude/2.1.118',
    'wakatime/v1.107.0 (linux-6.11.8) go1.23.3 Claude/2.1.118 jetbrains/PyCharm/2023.1',
    'wakatime/v1.115.2 (windows-10.0.22631.5335-x86_64) go1.24.2 Claude/unknown windows-wakatime/2.1.6',
    'wakatime/v1.130.1 (linux-6.6.87.2-microsoft-standard-WSL2-x86_64) go1.24.4 claude-code-wakatime/2.1.0',
    'wakatime/v1.131.0 (darwin-25.0.0-arm64) go1.24.4 Claude/0.11.3-0.11.3 macos-wakatime/5.27.2',
]


def random_user_agent() -> str:
    return random.choice(USER_AGENTS)


def weighted_type() -> str:
    """Pick an entity type, weighted so 'file' dominates as in real traffic."""
    types, weights = zip(*TYPE_WEIGHTS)
    return random.choices(types, weights=weights, k=1)[0]


def random_category(entity_type: str) -> Union[str, None]:
    """Pick a realistic category for the given entity type (None = unset)."""
    cat = random.choice(TYPE_CATEGORIES[entity_type])
    return cat if cat != '' else None


def random_entity(entity_type: str, project: str, language: str) -> str:
    """Generate a realistic entity string for the given type."""
    if entity_type == 'file':
        ext = LANGUAGES.get(language, 'txt')
        return f'/home/me/dev/{project}/{randomword(random.randint(2, 8))}.{ext}'
    if entity_type == 'domain':
        return random.choice(DOMAINS)
    if entity_type == 'url':
        org = random.choice(['acme', 'myteam', 'platform', 'infra', 'frontend'])
        repo = random.choice(['web-app', 'api', 'service', 'cli', 'docs'])
        return random.choice(URLS).format(org=org, repo=repo)
    if entity_type == 'app':
        return random.choice(APPS)
    if entity_type == 'event':
        return random.choice(EVENTS)
    return randomword(8)


class Heartbeat:
    def __init__(
            self,
            entity: str,
            project: str,
            language: str,
            time: float,
            is_write: bool = True,
            branch: str = 'master',
            type: str = 'file',
            user_agent: str = '',
            category: Union[str, None] = None
    ):
        self.entity: str = entity
        self.project: str = project
        self.language: str = language
        self.time: float = time
        self.is_write: bool = is_write
        self.branch: str = branch
        self.type: str = type
        self.category: Union[str, None] = category
        self.user_agent: str = user_agent


class ConfigParams:
    def __init__(self):
        self.api_url = ''
        self.api_key = ''
        self.n = 0
        self.n_projects = 0
        self.offset = 0
        self.seed = 0
        self.batch = False


def generate_data(n: int, n_projects: int = 5, n_past_hours: int = 24) -> List[Heartbeat]:
    data: List[Heartbeat] = []
    now: datetime = datetime.today()
    if n_projects >= len(PROJECTS):
        projects: List[str] = list(PROJECTS)
    else:
        projects = random.sample(PROJECTS, n_projects)
    languages: List[str] = list(LANGUAGES.keys())

    for _ in range(n):
        p: str = random.choice(projects)
        t: str = weighted_type()
        l: str = random.choice(languages) if t == 'file' else None
        b: str = random.choice(BRANCHES) if t == 'file' else ''
        delta: timedelta = timedelta(
            hours=random.randint(0, n_past_hours - 1),
            minutes=random.randint(0, 59),
            seconds=random.randint(0, 59),
            milliseconds=random.randint(0, 999),
            microseconds=random.randint(0, 999)
        )

        data.append(Heartbeat(
            entity=random_entity(t, p, l),
            project=p,
            language=l if (l and '?' not in l) else None,
            branch=b,
            time=(now - delta).timestamp(),
            user_agent=random_user_agent(),
            is_write=(t == 'file') and random.choice([True, False]),
            type=t,
            category=random_category(t),
        ))

    return data


def post_data_sync(data: List[Heartbeat], url: str, api_key: str):
    encoded_key: str = str(base64.b64encode(api_key.encode('utf-8')), 'utf-8')

    client = httpx.Client()
    response = client.post(url, json=[h.__dict__ for h in data], headers={
        'User-Agent': random_user_agent(),
        'Authorization': f'Basic {encoded_key}',
        'X-Machine-Name': MACHINE,
    })
    response.raise_for_status()


def make_gui(callback: Callable[[ConfigParams, Callable[[int], None]], None]) -> ('QApplication', 'QWidget'):
    # https://doc.qt.io/qt-6/qtwidgets-module.html
    from PyQt6.QtCore import Qt
    from PyQt6.QtWidgets import QApplication, QWidget, QFormLayout, QHBoxLayout, QVBoxLayout, QGroupBox, QLabel, \
        QLineEdit, QSpinBox, QProgressBar, QPushButton, QCheckBox, QMessageBox

    # Main app
    app = QApplication([])

    window = QWidget()
    window.setWindowTitle('Wakapi Sample Data Generator')
    window.setFixedSize(window.sizeHint())
    window.setMinimumWidth(350)

    container_layout = QVBoxLayout()

    # Top Controls
    form_layout_1 = QFormLayout()

    url_input_label = QLabel('URL:')
    url_input = QLineEdit()
    url_input.setPlaceholderText('Wakatime API Url')
    url_input.setText('http://localhost:3000/api')

    api_key_input_label = QLabel('API Key:')
    api_key_input = QLineEdit()
    api_key_input.setPlaceholderText(f'{"x" * 8}-{"x" * 4}-{"x" * 4}-{"x" * 4}-{"x" * 12}')

    form_layout_1.addRow(url_input_label, url_input)
    form_layout_1.addRow(api_key_input_label, api_key_input)

    # Middle controls
    form_layout_2 = QFormLayout()
    params_container = QGroupBox('Parameters')
    params_container.setLayout(form_layout_2)

    heartbeats_input_label = QLabel('# Heartbeats')
    heartbeats_input = QSpinBox()
    heartbeats_input.setMaximum(2147483647)
    heartbeats_input.setValue(100)

    projects_input_label = QLabel('# Projects:')
    projects_input = QSpinBox()
    projects_input.setMinimum(1)
    projects_input.setMaximum(len(PROJECTS))
    projects_input.setValue(5)

    offset_input_label = QLabel('Time Offset (hrs):')
    offset_input = QSpinBox()
    offset_input.setMinimum(-2147483647)
    offset_input.setMaximum(0)
    offset_input.setValue(-12)

    seed_input_label = QLabel('Random Seed:')
    seed_input = QSpinBox()
    seed_input.setMaximum(2147483647)
    seed_input.setValue(1337)

    batch_checkbox = QCheckBox('Batch Mode')
    batch_checkbox.setTristate(False)

    form_layout_2.addRow(heartbeats_input_label, heartbeats_input)
    form_layout_2.addRow(projects_input_label, projects_input)
    form_layout_2.addRow(offset_input_label, offset_input)
    form_layout_2.addRow(seed_input_label, seed_input)
    form_layout_2.addRow(batch_checkbox)

    # Bottom controls
    bottom_layout = QHBoxLayout()

    start_button = QPushButton('Generate')
    progress_bar = QProgressBar()
    progress_bar.setValue(0)

    bottom_layout.addWidget(progress_bar)
    bottom_layout.addWidget(start_button)

    # Wiring up
    container_layout.addLayout(form_layout_1)
    container_layout.addWidget(params_container)
    container_layout.addLayout(bottom_layout)
    container_layout.setStretch(1, 1)

    window.setLayout(container_layout)

    # Logic
    def parse_params() -> ConfigParams:
        params = ConfigParams()
        params.api_url = url_input.text()
        params.api_key = api_key_input.text()
        params.n = heartbeats_input.value()
        params.n_projects = projects_input.value()
        params.offset = offset_input.value()
        params.seed = seed_input.value()
        params.batch = batch_checkbox.isChecked()
        return params

    def update_progress(inc=1):
        current = progress_bar.value()
        updated = current + inc
        progress_bar.setValue(updated)
        if updated == progress_bar.maximum():
            progress_bar.setValue(0)
            start_button.setEnabled(True)

            dlg = QMessageBox()
            dlg.setWindowTitle('Success')
            dlg.setText('Done')
            dlg.exec()

            return

    def on_error(e):
        dlg = QMessageBox()
        dlg.setWindowTitle('Error')
        dlg.setText(e)
        btn = dlg.exec()
        start_button.setEnabled(True)

    def call_back():
        params = parse_params()
        progress_bar.setMaximum(params.n)
        progress_bar.setValue(0)
        start_button.setEnabled(False)
        callback(params, update_progress, on_error)

    start_button.clicked.connect(call_back)

    return app, window


def projects_count(value: str) -> int:
    n = int(value)
    if n < 1 or n > len(PROJECTS):
        raise argparse.ArgumentTypeError(f'must be between 1 and {len(PROJECTS)}')
    return n


def parse_arguments():
    parser = argparse.ArgumentParser(description='Wakapi test data insertion script.')
    parser.add_argument('--headless', default=False, help='do not show a gui', action='store_true')
    parser.add_argument('-n', type=int, default=20, help='total number of random heartbeats to generate and insert')
    parser.add_argument('-u', '--url', type=str, default='http://localhost:3000/api', help='url of your api\'s heartbeats endpoint')
    parser.add_argument('-k', '--apikey', type=str, required=True, help='your api key (to get one, go to the web interface, create a new user, log in and copy the key)')
    parser.add_argument('-p', '--projects', type=projects_count, default=5, help=f'number of different fake projects to generate (1-{len(PROJECTS)})')
    parser.add_argument('-o', '--offset', type=int, default=24, help='negative time offset in hours from now for to be used as an interval within which to generate heartbeats for')
    parser.add_argument('-s', '--seed', type=int, default=2020, help='a seed for initializing the pseudo-random number generator')
    parser.add_argument('-b', '--batch', default=False, help='batch mode (push all heartbeats at once)', action='store_true')
    return parser.parse_args()


def args_to_params(parsed_args: argparse.Namespace) -> (ConfigParams, bool):
    params = ConfigParams()
    params.n = parsed_args.n
    params.n_projects = parsed_args.projects
    params.offset = parsed_args.offset
    params.seed = parsed_args.seed
    params.api_url = parsed_args.url
    params.api_key = parsed_args.apikey
    params.batch = parsed_args.batch
    return params, not parsed_args.headless


def randomword(length: int) -> str:
    letters = string.ascii_lowercase + 'äöü'  # test utf8 characters as well
    return ''.join(random.choice(letters) for _ in range(length))


def run(params: ConfigParams, update_progress: Callable[[int], None], on_error: Callable[[str], None]):
    random.seed(params.seed)
    data: List[Heartbeat] = generate_data(
        params.n,
        params.n_projects,
        params.offset * -1 if params.offset < 0 else params.offset
    )

    # batch-mode won't work when using sqlite backend
    try:
        if params.batch:
            post_data_sync(data, f'{params.api_url}/heartbeats', params.api_key)
            update_progress(len(data))
        else:
            for d in data:
                post_data_sync([d], f'{params.api_url}/heartbeats', params.api_key)
                update_progress(1)
    except RequestError as e:
        on_error(str(e))


if __name__ == '__main__':
    params, show_gui = args_to_params(parse_arguments())
    if show_gui:
        app, window = make_gui(callback=run)
        window.show()
        app.exec()
    else:
        from tqdm import tqdm

        pbar = tqdm(total=params.n)
        run(params, pbar.update, print)
