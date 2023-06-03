import torch
import subprocess
import whisper
import requests
import subprocess
from pyannote.audio import Pipeline
from pyannote_whisper.utils import diarize_text
from pywhispercpp.model import Model

torch.cuda.is_available()
pipeline = Pipeline.from_pretrained("pyannote/speaker-diarization@2.1",
                                    use_auth_token="TOKEN")
devices = torch.device("cuda:0") 

import click 
@click.command()
@click.argument("path", type=click.Path(exists=True))
@click.argument("GUID")

def file_handler(path, guid):
    try:
        file_wav = path[:-4] + ".wav"
        print(path[-4:])

        if path[-4:] == ".mp4": 
            subprocess.call(['ffmpeg', '-i', path, '-vn', '-acodec', 'pcm_s16le', '-ar', '44100', '-ac', '2', file_wav])  
        elif path[-4:] == ".mp3":
            subprocess.call(['ffmpeg', '-i', path, file_wav])
        elif path[-4:] == ".mkv":
            subprocess.call(['ffmpeg', '-i', path, '-vn', '-acodec', 'pcm_s16le', '-ar', '44100', '-ac', '2', file_wav])  
        else:
            file_wav = path[:-5] + ".wav"
            print(path + "\n" + file_wav + "\n")
            subprocess.call(['ffmpeg', '-i', path, '-vn', file_wav])

        model_size = "large" 
        model = whisper.load_model(f"{model_size}", device = devices)

        asr_result = model.transcribe(file_wav)
        diarization_result = pipeline(file_wav)
        final_result = diarize_text(asr_result, diarization_result)
        file_mp3 = file_wav.replace(".wav", ".mp3")
        f = open(f"{file_mp3}_text.txt", "a", encoding="utf-8")

        for seg, spk, sent in final_result:
            line = f'{seg.start:.2f} {seg.end:.2f} {spk} {sent}'
            print(line)
            f.write(line + "\n")
        f.close()

        payload = {'file_id': guid, 'error': ''}
        requests.get('http://server/whisper_ping', params=payload)
    except Exception as error:
        payload = {'file_id': guid, 'error': str(error)}
        requests.get('http://server/whisper_ping', params=payload)

if __name__ == '__main__':
    file_handler()