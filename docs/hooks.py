"""MkDocs build hooks."""

from datetime import datetime

COPYRIGHT_START_YEAR = 2021


def on_config(config):
    """Set copyright end year from the current calendar year at build time."""
    year = datetime.now().year
    if year > COPYRIGHT_START_YEAR:
        config.copyright = (
            f"© EDENLAB. ALL RIGHTS RESERVED, {COPYRIGHT_START_YEAR} - {year}"
        )
    else:
        config.copyright = f"© EDENLAB. ALL RIGHTS RESERVED, {COPYRIGHT_START_YEAR}"
    return config
