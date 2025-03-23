const SetupProfile = ({ open, handleClose }) => {
  const [hashnodePAT, setHashnodePAT] = useState('');
  const [openAIKey, setOpenAIKey] = useState('');

  const handleSetup = () => {
    console.log('Setup info:', { hashnodePAT, openAIKey });
  };

  return (
    <Modal open={open} onClose={handleClose}>
      <Box
        sx={{
          width: 400,
          padding: '2rem',
          backgroundColor: '#2E2E2E',
          borderRadius: '12px',
          color: '#FFFFFF',
          textAlign: 'center',
        }}
      >
        <Typography
          variant="h5"
          sx={{ fontWeight: 'bold', marginBottom: '1rem', color: '#FF6B6B' }}
        >
          Complete Your Setup
        </Typography>
        <TextField
          label="Hashnode Personal Access Token"
          variant="outlined"
          value={hashnodePAT}
          onChange={(e) => setHashnodePAT(e.target.value)}
          fullWidth
          sx={{
            marginBottom: 2,
            backgroundColor: '#3A3A3A',
            borderRadius: '5px',
            input: { color: '#FFFFFF' },
          }}
        />
        <TextField
          label="OpenAI API Key"
          variant="outlined"
          value={openAIKey}
          onChange={(e) => setOpenAIKey(e.target.value)}
          fullWidth
          sx={{
            marginBottom: 2,
            backgroundColor: '#3A3A3A',
            borderRadius: '5px',
            input: { color: '#FFFFFF' },
          }}
        />
        <Button
          variant="contained"
          fullWidth
          sx={{ marginTop: 2, backgroundColor: '#FF6B6B', color: 'white', fontWeight: 'bold' }}
          onClick={handleSetup}
        >
          Save
        </Button>
        <Typography variant="body2" sx={{ marginTop: 2, color: '#B8B8B8', cursor: 'pointer' }}>
          Connect X Account
        </Typography>
        <Typography variant="body2" sx={{ marginTop: 1, color: '#B8B8B8', cursor: 'pointer' }}>
          Connect LinkedIn Account
        </Typography>
      </Box>
    </Modal>
  );
};

export default SetupProfile;
