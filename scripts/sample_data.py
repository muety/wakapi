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
UA = 'wakatime/13.0.7 (Linux-4.15.0-91-generic-x86_64-with-glibc2.4) Python3.8.0.final.0 generator/1.42.1 generator-wakatime/4.0.0'
LANGUAGES = {
    'Go': 'go',
    'Java': 'java',
    'JavaScript': 'js',
    'Python': 'py',
    'PHP': 'php',
    'Blade': 'blade.php',  # https://github.com/muety/wakapi/issues/172
    '?': 'astro',  # simulate language unknown to wakatime-cli
}
BRANCHES = ['master', 'feature-1', 'feature-2']


class Heartbeat:
    def __init__(
            self,
            entity: str,
            project: str,
            language: str,
            time: float,
            is_write: bool = True,
            branch: str = 'master',
            type: str = 'file'
    ):
        self.entity: str = entity
        self.project: str = project
        self.language: str = language
        self.time: float = time
        self.is_write: bool = is_write
        self.branch: str = branch
        self.type: str = type
        self.category: Union[str, None] = None


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
    projects: List[str] = [randomword(random.randint(5, 10)) for _ in range(n_projects)]
    languages: List[str] = list(LANGUAGES.keys())

    for _ in range(n):
        p: str = random.choice(projects)
        l: str = random.choice(languages)
        f: str = randomword(random.randint(2, 8))
        b: str = random.choice(BRANCHES)
        delta: timedelta = timedelta(
            hours=random.randint(0, n_past_hours - 1),
            minutes=random.randint(0, 59),
            seconds=random.randint(0, 59),
            milliseconds=random.randint(0, 999),
            microseconds=random.randint(0, 999)
        )

        data.append(Heartbeat(
            entity=f'/home/me/dev/{p}/{f}.{LANGUAGES[l]}',
            project=p,
            language=l if not '?' in l else None,
            branch=b,
            time=(now - delta).timestamp()
        ))

    return data


def post_data_sync(data: List[Heartbeat], url: str, api_key: str):
    encoded_key: str = str(base64.b64encode(api_key.encode('utf-8')), 'utf-8')

    client = httpx.Client()
    response = client.post(url, json=[h.__dict__ for h in data], headers={
        'User-Agent': UA,
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


def parse_arguments():
    parser = argparse.ArgumentParser(description='Wakapi test data insertion script.')
    parser.add_argument('--headless', default=False, help='do not show a gui', action='store_true')
    parser.add_argument('-n', type=int, default=20, help='total number of random heartbeats to generate and insert')
    parser.add_argument('-u', '--url', type=str, default='http://localhost:3000/api', help='url of your api\'s heartbeats endpoint')
    parser.add_argument('-k', '--apikey', type=str, required=True, help='your api key (to get one, go to the web interface, create a new user, log in and copy the key)')
    parser.add_argument('-p', '--projects', type=int, default=5, help='number of different fake projects to generate')
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
    letters = string.ascii_lowercase + 'äöü💩'  # test utf8 and utf8mb4 characters as well
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